package main

import (
	"encoding/json"
	"fmt"
	"github.com/Inspur-Data/ipamwrapper/api/v1/models"
	"github.com/Inspur-Data/ipamwrapper/cmd/router/bin"
	"github.com/containernetworking/cni/pkg/skel"
)

func main() {
	//time.Sleep(1*time.Second)
	fmt.Sprintf("test")
	bin.CmdAdd(ipamConfig())
}

func ipamConfig() *skel.CmdArgs {
	const (
		cniVersion = "0.3.1"
		netName    = "net1"
	)
	podNamespace := "podNamespace"
	podName := "podName"
	routerRules := []*models.RouterRule{
		{
			Dst: "10.0.0.0/24",
			Gw: "10.0.0.1",
		},
		{
			Dst: "10.1.0.0/24",
			Gw: "10.1.0.1",
		},
	}
	serviceCIDR := []string{"10.2.0.2/24","10.3.0.2/24"}
	routerConfig := &models.RouterConfig{
		DetectGateway:                false,
		PodDefaultRouteNIC:          "eth0",
		Routes:                      routerRules,
		ServiceCIDR:                serviceCIDR,
		TableName:                  "main",
	}

	netConf := NetConf{
		CNIVersion:   cniVersion,
		Name:         "router",
		RouterConfig: routerConfig,
	}
	stdinData,_ := json.Marshal(netConf)
	args := &skel.CmdArgs{
		ContainerID: fmt.Sprintf("dummy-%d", 1),
		Netns:       "/some/router",
		IfName:      "eth0",
		StdinData:   stdinData,
		Args:        cniArgs(podNamespace, podName),
	}
	fmt.Sprintf("args: %v",args)
	return args
}
type NetConf struct {
	// detect gateway
	Name string `json:"name"`
	CNIVersion string      `json:"cniVersion"`
	Type string `json:"type"`
	RouterConfig *models.RouterConfig  `json:"routerConfig"`
}
func cniArgs(podNamespace string, podName string) string {
	return fmt.Sprintf("IgnoreUnknown=1;K8S_POD_NAMESPACE=%s;K8S_POD_NAME=%s", podNamespace, podName)
}