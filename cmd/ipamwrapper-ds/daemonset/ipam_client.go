package daemonset

import (
	"context"
	"github.com/Inspur-Data/ipamwrapper/api/v1/client"
	"github.com/Inspur-Data/ipamwrapper/api/v1/server/restapi"
	"github.com/Inspur-Data/ipamwrapper/pkg/ipam"
	"github.com/Inspur-Data/ipamwrapper/pkg/manager/endpointmanager"
	"github.com/Inspur-Data/ipamwrapper/pkg/manager/ippoolmanager"
	"github.com/Inspur-Data/ipamwrapper/pkg/manager/nsmanager"
	"github.com/Inspur-Data/ipamwrapper/pkg/manager/podmanager"
	"go.uber.org/atomic"
	"net/http"
	ctrl "sigs.k8s.io/controller-runtime"
)

type IPAMAgent struct {
	Cfg Config

	// InnerCtx is the context that can be used during shutdown.
	// It will be cancelled after receiving an interrupt or termination signal.
	InnerCtx    context.Context
	InnerCancel context.CancelFunc

	// manager
	Mgr         ctrl.Manager
	IPAM        ipam.IPAM
	IPPoolMgr   ippoolmanager.IPPoolManager
	EndpointMgr endpointmanager.EndpointManager
	NSMgr       nsmanager.NsManager
	PodMgr      podmanager.PodManager

	// handler
	HttpServer        *restapi.Server
	UnixServer        *restapi.Server
	MetricsHttpServer *http.Server

	// client
	unixClient *client.IpamwrapperAgentAPI

	// probe
	IsStartupProbe atomic.Bool
}

var ipamAgent = new(IPAMAgent)
