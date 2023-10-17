package daemonset

import (
	"context"
	ipam2 "github.com/Inspur-Data/ipamwrapper/pkg/ipam"
	"github.com/Inspur-Data/ipamwrapper/pkg/logging"
	"time"
)

func Daemon() {
	//set the context
	ipamAgent.InnerCtx, ipamAgent.InnerCancel = context.WithCancel(context.Background())

	//new manager
	mgr, err := newManager()
	if err != nil {
		logging.Panicf("new manger failed:%v", err)
	}
	ipamAgent.Mgr = mgr
	ipamAgent.Cfg = ConfigInstance
	//init manager
	initManager()

	ipam, err := ipam2.NewIPAM(
		ipam2.IPAMConfig{
			EnableIPv4:               ipamAgent.Cfg.EnableIPv4,
			EnableIPv6:               ipamAgent.Cfg.EnableIPv6,
			ClusterDefaultIPv4IPPool: ipamAgent.Cfg.ClusterDefaultIPv4IPPool,
			ClusterDefaultIPv6IPPool: ipamAgent.Cfg.ClusterDefaultIPv6IPPool,
			OperationRetries:         ipamAgent.Cfg.WaitSubnetPoolMaxRetries,
			OperationGapDuration:     time.Duration(ipamAgent.Cfg.WaitSubnetPoolTime) * time.Second,
			IPv4ReservedIP:           ipamAgent.Cfg.IPv4ReservedIPs,
			IPv6ReservedIP:           ipamAgent.Cfg.IPv6ReservedIPs,
			EnableSpiderSubnet:       true,
			EnableStatefulSet:        true,
		},
		ipamAgent.PodMgr,
		ipamAgent.EndpointMgr,
		ipamAgent.IPPoolMgr,
		ipamAgent.NSMgr,
		ipamAgent.StsMgr,
		ipamAgent.NodeMgr,
	)
	ipamAgent.IPAM = ipam
	go func() {
		err := mgr.Start(ipamAgent.InnerCtx)
		if err != nil {
			logging.Errorf("manager start failed:%v", err)
		}
	}()

	//init http client
	httpClient, err := NewAgentOpenAPIHttpClient("localhost:" + ipamAgent.Cfg.HealthHttpPort)
	if nil != err {
		logging.Errorf("failed to create agent client: %v", err)
	}
	ipamAgent.httpClient = httpClient
}
