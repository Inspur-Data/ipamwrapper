package daemonset

import (
	inspuripamv1 "github.com/Inspur-Data/ipamwrapper/pkg/k8s/api/v1"
	"github.com/Inspur-Data/ipamwrapper/pkg/logging"
	"github.com/Inspur-Data/ipamwrapper/pkg/manager/endpointmanager"
	"github.com/Inspur-Data/ipamwrapper/pkg/manager/ippoolmanager"
	"github.com/Inspur-Data/ipamwrapper/pkg/manager/nsmanager"
	"github.com/Inspur-Data/ipamwrapper/pkg/manager/podmanager"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
)

var scheme = runtime.NewScheme()

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(inspuripamv1.AddToScheme(scheme))
}
func newManager() (ctrl.Manager, error) {
	config := ctrl.GetConfigOrDie()
	config.Burst = 100
	config.QPS = 50

	mgr, err := ctrl.NewManager(config, ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     "0",
		HealthProbeBindAddress: "0",
	})
	if err != nil {
		logging.Errorf("newmanager failed:%v", err)
		return nil, err
	}

	/*
		if err := mgr.GetFieldIndexer().IndexField(ipamAgent.InnerCtx, &inspuripamv1.IPPool{}, "spec.default", func(raw client.Object) []string {
			ipPool := raw.(*inspuripamv1.IPPool)
			return []string{strconv.FormatBool(*ipPool.Spec.Default)}
		}); err != nil {
			return nil, err
		}*/

	return mgr, nil
}
func initManager() {
	logging.Debugf("init namespace manger")
	nsManager, err := nsmanager.NewNamespaceManager(
		ipamAgent.Mgr.GetClient(),
		ipamAgent.Mgr.GetAPIReader(),
	)
	if err != nil {
		logging.Panicf("init namespace manager failed:%v", err)
	}
	ipamAgent.NSMgr = nsManager

	logging.Debugf("init pod manager")
	podManager, err := podmanager.NewPodManager(
		ipamAgent.Mgr.GetClient(),
		ipamAgent.Mgr.GetAPIReader(),
	)
	if err != nil {
		logging.Panicf("init pod manager failed:%v", err)
	}
	ipamAgent.PodMgr = podManager

	logging.Debugf("init endpoint manager")
	endpointManager, err := endpointmanager.NewEndpointManager(
		ipamAgent.Mgr.GetClient(),
		ipamAgent.Mgr.GetAPIReader(),
	)
	if err != nil {
		logging.Panicf("init endpoint manager failed:%v", err)
	}
	ipamAgent.EndpointMgr = endpointManager

	logging.Debugf("init ippool manager ")
	ipPoolManager, err := ippoolmanager.NewIPPoolManager(
		ippoolmanager.MgrConfig{
			MaxAllocatedIPs: &ipamAgent.Cfg.IPPoolMaxAllocatedIPs,
		},
		ipamAgent.Mgr.GetClient(),
		ipamAgent.Mgr.GetAPIReader(),
	)
	if err != nil {
		logging.Panicf("init ippool manager failed: %v", err)
	}
	ipamAgent.IPPoolMgr = ipPoolManager

}
