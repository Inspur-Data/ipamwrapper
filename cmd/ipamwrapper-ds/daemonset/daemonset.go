package daemonset

import (
	ipam2 "github.com/Inspur-Data/ipamwrapper/pkg/ipam"
	"github.com/Inspur-Data/ipamwrapper/pkg/logging"
	"time"
)

func Daemon() {
	//new manager
	mgr, err := newManager()
	if err != nil {
		logging.Panicf("new manger failed:%v", err)
	}

	ipamAgent.Mgr = mgr

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
		},
		ipamAgent.PodMgr,
		ipamAgent.EndpointMgr,
		ipamAgent.IPPoolMgr,
		ipamAgent.NSMgr,
	)
	ipamAgent.IPAM = ipam

	go func() {
		mgr.Start(ipamAgent.InnerCtx)
	}()

}
