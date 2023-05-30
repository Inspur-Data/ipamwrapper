package main

import (
	"github.com/Inspur-Data/k8-ipam/cmd/k8-ipam-ds/daemonset"
	"github.com/Inspur-Data/k8-ipam/pkg/logging"
	"net/http"
)

func main() {
	srv, err := daemonset.NewAgentOpenAPIUnixServer()
	if err != nil {
		logging.Errorf("get unix server instance failed:%v", err)
		return
	}

	go func() {
		logging.Debugf("start a k8ipam unix server")
		err := srv.Serve()
		if err != nil {
			if err == http.ErrServerClosed {
				return
			}
			logging.Panicf("start a k8ipam unix server failed: %v", err)
		}
	}()
}
