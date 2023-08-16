package bin

import (
	"fmt"
	"github.com/Inspur-Data/ipamwrapper/api/v1/models"
	"github.com/Inspur-Data/ipamwrapper/pkg/config"
	"github.com/Inspur-Data/ipamwrapper/pkg/logging"
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	current "github.com/containernetworking/cni/pkg/types/100"
	//"time"
)

func CmdDel(args *skel.CmdArgs) (err error) {
	// for calu consuming  time
	//startTime := time.Now()

	// parse args
	podArgs := config.PodArgs{}
	if err = types.LoadArgs(args.Args, &podArgs); nil != err {
		return fmt.Errorf("failed to load CNI ENV args: %w", err)
	}

	// parse router config info
	routerConfig := models.RouterConfig{}
	conf, err := ParseConfig(args.StdinData, &routerConfig)
	if err != nil || conf == nil{
		return err
	}

	// parse prevResult
	prevResult, err := current.GetResult(conf.PrevResult)
	logging.Errorf("prevResult error info: %v", prevResult)
	if err != nil {
		//logger.Error("failed to convert prevResult", zap.Error(err))
		logging.Errorf("prevResult error info: %v", err)
		return err
	}
	/*
	// validate prevResult err
	ipFamily, err := ipamip.GetIPFamilyByResult(prevResult)
	if err != nil {
		logging.Errorf("failed to GetIPFamilyByResult", zap.Error(err))
		return err
	}
    */

	/*netns, err := ns.GetNS(args.Netns)
	if err != nil {
		logging.Errorf(err.Error())
		return fmt.Errorf("failed to GetNS %q: %v", args.Netns, err)
	}
	defer netns.Close()*/

	// set rule in ns
	/*for  _, route := range routerConfig.Routes{
		// get dst and gw
		_,dst,err := net.ParseCIDR(route.Dst)
		if err != nil {
			return fmt.Errorf("failed to translate dst :%v,err: %v",dst, err.Error())
		}
		//DelRoute(100, ipFamily,netlink.SCOPE_UNIVERSE,args.IfName,dst,net.ParseIP(route.Gw))
	}*/

	return nil
}
/*
// DelRoute in ns
func DelRoute(ruleTable, ipFamily int, scope netlink.Scope, iface string, dst *net.IPNet, gw net.IP) error {
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
	if err = netlink.RouteDel(route); err != nil && !os.IsExist(err) {
		logging.Errorf("failed to RouteDel,route:%v ,err:%v",route.String(),err.Error())
		return fmt.Errorf("failed to del route table(%v): %v", route.String(), err)
	}
	return nil
}
*/