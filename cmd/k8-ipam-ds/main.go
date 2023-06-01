package main

import (
	"github.com/Inspur-Data/k8-ipam/cmd/k8-ipam-ds/daemonset"
	"github.com/Inspur-Data/k8-ipam/pkg/logging"
	"net/http"
)

func main() {
	logging.Debugf("http server will start.....")
	srv, err := daemonset.NewAgentOpenAPIUnixServer()
	// srv, err := daemonset.NewIPAMHttpServer()
	if err != nil {
		logging.Errorf("get unix server instance failed:%v", err)
		return
	}

	logging.Debugf("start a k8ipam unix server")
	err = srv.Serve()
	if err != nil {
		if err == http.ErrServerClosed {
			return
		}
		logging.Panicf("start a k8ipam unix server failed: %v", err)
	}
}

func init() {
	err := daemonset.ParseConfiguration()
	if err != nil {
		logging.Panicf("ParseConfig failed: %v", err)
	}

	err = daemonset.LoadConfigmap()
	if err != nil {
		logging.Panicf("Loadconfigmap failed: %v", err)
	}
}
