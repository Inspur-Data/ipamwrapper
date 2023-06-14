// Copyright 2022 Authors of Inspur
// SPDX-License-Identifier: Apache-2.0
package daemonset

import (
	"strconv"

	k8IPAMClient "github.com/Inspur-Data/ipamwrapper/api/v1/client"
	k8IPAMRestapi "github.com/Inspur-Data/ipamwrapper/api/v1/server/restapi"
	k8IPAMOperation "github.com/Inspur-Data/ipamwrapper/api/v1/server/restapi/operations"
	"github.com/go-openapi/loads"
)

// NewIPAMHttpServer create a new instance of the IPAM server on the http.
func NewIPAMHttpServer() (*k8IPAMRestapi.Server, error) {
	// read yaml spec
	swaggerSpec, err := loads.Embedded(k8IPAMRestapi.SwaggerJSON, k8IPAMRestapi.FlatSwaggerJSON)
	if nil != err {
		return nil, err
	}

	// create new service API
	api := k8IPAMOperation.NewK8IpamAgentAPIAPI(swaggerSpec)

	// set runtime Handler
	api.K8IpamAgentPostIpamHandler = unixPostIp
	api.K8IpamAgentDeleteIpamHandler = unixDeleteIp

	// new agent OpenAPI server with api
	srv := k8IPAMRestapi.NewServer(api)

	// k8ipam daemonset owns Unix server and Http server, the Unix server uses for interaction, and the Http server uses for K8s or CLI command.
	// In config file openapi.yaml, we already set x-schemes with value 'unix', so we need set Http server's listener with value 'http'.
	srv.EnabledListeners = k8IPAMClient.DefaultSchemes
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
