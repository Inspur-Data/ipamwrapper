package bin

import (
	"encoding/json"
	"fmt"
	"github.com/Inspur-Data/ipamwrapper/api/v1/models"
	"github.com/Inspur-Data/ipamwrapper/pkg/logging"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/vishvananda/netlink"
	"os"

	"github.com/containernetworking/cni/pkg/version"
	//"github.com/vishvananda/netns"
	//"go.uber.org/zap"

	//"github.com/vishvananda/netlink"
	//"go.uber.org/zap"
	"k8s.io/utils/pointer"
	"net"
	//"time"
	"github.com/Inspur-Data/ipamwrapper/pkg/config"
	//ipamip "github.com/Inspur-Data/ipamwrapper/pkg/ip"
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
)

var (
	defaultOverlayVethName = "eth0"
)

type RouterConfig struct {
	types.NetConf
	TableName          string       `json:"detectGateway,omitempty"`
	DetectGateway      *bool        `json:"detectGateway,omitempty"`
	ServiceCIDR        []string     `json:"serviceCIDR,omitempty"`
	PodDefaultRouteNIC string       `json:"podDefaultRouteNic,omitempty"`
	Routes             []RouterInfo `json:"routes"`
}
type RouterInfo struct {
	Dst string `json:"dst"`
	Gw  string `json:"dw"`
}

func CmdAdd(args *skel.CmdArgs) (err error) {
	// for calu consuming  time
	//startTime := time.Now()
	podArgs := config.PodArgs{}
	if err = types.LoadArgs(args.Args, &podArgs); nil != err {
		return fmt.Errorf("failed to load CNI ENV args: %w", err)
	}
	logging.Debugf("podArgs :%v", podArgs)

	routerNetConf := config.RouterNetConf{}
	if err = json.Unmarshal(args.StdinData, &routerNetConf); err != nil {
		return fmt.Errorf("failed to routerConfig : %v", err)
	}
	logging.Debugf("routerConfig :%v", routerNetConf)

	netns, err := ns.GetNS(args.Netns)
	if err != nil {
		return logging.Errorf("failed to GetNS %q: %v", args.Netns, err)
	}
	defer netns.Close()
	logging.Debugf("netns :%v", netns)

	routerConfig := routerNetConf.RouterConfig
	logging.Debugf("netns :%v", netns)
	//globalDynamicRouterGW := routerConfig.DynamicRouterGW
	err = netns.Do(func(_ ns.NetNS) error {
		var v4Gw net.IP
		//var v6Gw net.IP

		logging.Debugf("ifName :%v", args.IfName)
		// get  ip by interface name
		link, err := netlink.LinkByName(args.IfName)
		if err != nil {
			return logging.Errorf("Failed to get link:", err.Error())
		}

		// 获取 IP 地址列表
		addrs, err := netlink.AddrList(link, netlink.FAMILY_V4)
		if err != nil {
			return logging.Errorf("Failed to get address list:", err.Error())
		}

		if len(addrs) == 0 {
			logging.Debugf("ifName's ip is empty :%v", args.IfName)
			return nil
		}

		podArgs.IP = addrs[0].IP
		logging.Debugf("podArgs IP :%v", podArgs.IP)

		routers, err := netlink.RouteGet(podArgs.IP)
		if err != nil {
			return fmt.Errorf("failed to RouteGet Pod IP(%s): %v", podArgs.IP, err)
		}

		if len(routers) == 0 {
			logging.Debugf("podName-pod:%v-%v routers is empty", podArgs.K8S_POD_NAMESPACE, podArgs.K8S_POD_NAME)
			return nil
		}

		if podArgs.IP.To4() != nil && v4Gw == nil {
			v4Gw = routers[0].Src
		}
		/*if podArgs.IP.To4() == nil && v6Gw == nil {
			v6Gw = routes[0].Src
		}*/

		// set rule in ns
		for _, route := range routerConfig.Routes {
			// get dst and gw
			_, dst, err := net.ParseCIDR(route.V4Dst)
			if err != nil {
				return fmt.Errorf("failed to translate dst :%v,err: %v", dst, err.Error())
			}
			logging.Debugf("dst :%v", dst)

			dynamicRouterGW := route.DynamicRouterGW

			// v4 gw info  from nad router info
			if !dynamicRouterGW {
				v4Gw = net.ParseIP(route.V4Gw)
			}
			logging.Debugf(" gw :%v", v4Gw)

			err = AddRoute(100, netlink.FAMILY_V4, netlink.SCOPE_UNIVERSE, args.IfName, dst, v4Gw)
			// we skip over duplicate routes as we assume the first one wins
			if !os.IsExist(err) {
				return logging.Errorf("failed to addRoute: %v", err.Error())
			}
		}
		return nil
	})
	return nil
}

// ParseConfig parses the supplied configuration (and prevResult) from stdin.
func ParseConfig(stdin []byte, routerConfig *models.RouterConfig) (*RouterConfig, error) {
	var err error
	conf := RouterConfig{}

	if err = json.Unmarshal(stdin, &conf); err != nil {
		return nil, fmt.Errorf("failed to parse config: %v", err)
	}

	if err = version.ParsePrevResult(&conf.NetConf); err != nil {
		return nil, fmt.Errorf("failed to parse prevResult: %v", err)
	}
	/*
		if err = routerConfig.Validate(strfmt.Default); err != nil {
			return nil, err
		}
	*/
	if conf.PodDefaultRouteNIC == "" {
		conf.PodDefaultRouteNIC = defaultOverlayVethName
	}

	if conf.DetectGateway == nil {
		conf.DetectGateway = pointer.Bool(routerConfig.DetectGateway)
	}

	if conf.PodDefaultRouteNIC == "" && routerConfig.PodDefaultRouteNIC != "" {
		conf.PodDefaultRouteNIC = routerConfig.PodDefaultRouteNIC
	}

	if len(conf.Routes) == 0 {
		// if not have routes,we don't show any error
		return nil, nil
	}
	return &conf, nil
}

// AddRoute add static route to specify rule table
func AddRoute(ruleTable, ipFamily int, scope netlink.Scope, iface string, dst *net.IPNet, gw net.IP) error {
	link, err := netlink.LinkByName(iface)
	if err != nil {
		logging.Errorf(err.Error())
		return err
	}
	// todo handle  table
	route := &netlink.Route{
		LinkIndex: link.Attrs().Index,
		Scope:     scope,
		Dst:       dst,
		Gw:        gw,
		//Table:     ruleTable,
	}
	if err = netlink.RouteAdd(route); err != nil && !os.IsExist(err) {
		return logging.Errorf("failed to RouteAdd,route:%v ,err:%v", route.String(), err.Error())
	}
	return nil
}
