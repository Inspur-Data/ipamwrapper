// Copyright 2022 Authors of Inspur
// SPDX-License-Identifier: Apache-2.0
package daemonset

import (
	"strconv"

	IPAMClient "github.com/Inspur-Data/ipamwrapper/api/v1/client"
	IPAMRestapi "github.com/Inspur-Data/ipamwrapper/api/v1/server/restapi"
	IPAMOperation "github.com/Inspur-Data/ipamwrapper/api/v1/server/restapi/operations"
	"github.com/go-openapi/loads"
)

// NewIPAMHttpServer create a new instance of the IPAM server on the http.
func NewIPAMHttpServer() (*IPAMRestapi.Server, error) {
	// read yaml spec
	swaggerSpec, err := loads.Embedded(IPAMRestapi.SwaggerJSON, IPAMRestapi.FlatSwaggerJSON)
	if nil != err {
		return nil, err
	}

	// create new service API
	api := IPAMOperation.NewIpamwrapperAgentAPIAPI(swaggerSpec)

	// set runtime Handler
	api.IpamwrapperAgentPostIpamHandler = unixPostIp
	api.IpamwrapperAgentDeleteIpamHandler = unixDeleteIp

	// new agent OpenAPI server with api
	srv := IPAMRestapi.NewServer(api)

	// k8ipam daemonset owns Unix server and Http server, the Unix server uses for interaction, and the Http server uses for K8s or CLI command.
	// In config file openapi.yaml, we already set x-schemes with value 'unix', so we need set Http server's listener with value 'http'.
	srv.EnabledListeners = IPAMClient.DefaultSchemes
	//todo port need to set
	port, err := strconv.Atoi(ConfigInstance.HttpPort)
	if nil != err {
		return nil, err
	}
	srv.Port = port

	// configure API and handlers with some default values.
	srv.ConfigureAPI()

	return srv, nil
}
