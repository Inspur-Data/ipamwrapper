// Copyright 2022 Authors of Inspur
// SPDX-License-Identifier: Apache-2.0

package daemonset

import (
	"github.com/jessevdk/go-flags"

	k8IPAMRestapi "github.com/Inspur-Data/k8-ipam/api/v1/server/restapi"
	k8IPAMOperation "github.com/Inspur-Data/k8-ipam/api/v1/server/restapi/operations"
	"github.com/go-openapi/loads"
)

// NewAgentOpenAPIUnixServer instantiates a new instance of the agent OpenAPI server on the unix.
func NewAgentOpenAPIUnixServer() (*k8IPAMRestapi.Server, error) {
	// read yaml spec
	swaggerSpec, err := loads.Embedded(k8IPAMRestapi.SwaggerJSON, k8IPAMRestapi.FlatSwaggerJSON)
	if nil != err {
		return nil, err
	}

	// create new service API
	api := k8IPAMOperation.NewK8IpamAgentAPIAPI(swaggerSpec)

	// daemonset API
	api.K8IpamAgentPostIpamHandler = unixPostIp
	api.K8IpamAgentDeleteIpamHandler = unixDeleteIp
	// new agent OpenAPI server with api
	srv := k8IPAMRestapi.NewServer(api)

	// set spiderpool-agent Unix server with specified unix socket path.
	//todo set the unixSocketPath
	srv.SocketPath = flags.Filename("")

	// configure API and handlers with some default values.
	srv.ConfigureAPI()

	return srv, nil
}
