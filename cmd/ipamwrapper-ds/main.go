package main

import (
	"github.com/Inspur-Data/ipamwrapper/cmd/ipamwrapper-ds/daemonset"
	"github.com/Inspur-Data/ipamwrapper/pkg/logging"
	"net/http"
)

func main() {
	//init resource
	daemonset.Daemon()
	logging.Debugf("http server will start.....")
	//start the health_check httpserver
	healthSrv, err := daemonset.NewIPAMHealthHttpServer()
	if err == nil && healthSrv != nil {
		go func() {
			er := healthSrv.Serve()
			if err != nil {
				logging.Errorf("health check server start fialed:%v", er)
			}
		}()
	}

	srv, err := daemonset.NewAgentOpenAPIUnixServer()
	//srv, err := daemonset.NewIPAMHttpServer()
	if err != nil {
		logging.Errorf("get unix server instance failed:%v", err)
		return
	}
	err = srv.Serve()
	if err != nil {
		if err == http.ErrServerClosed {
			logging.Errorf("server has closed")
			return
		}
		logging.Panicf("start a ipam unix server failed: %v", err)
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
