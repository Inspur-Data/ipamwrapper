// Copyright 2022 Authors of Inspur
// SPDX-License-Identifier: Apache-2.0

package daemonset

import (
	"github.com/jessevdk/go-flags"

	IPAMRestapi "github.com/Inspur-Data/ipamwrapper/api/v1/server/restapi"
	IPAMOperation "github.com/Inspur-Data/ipamwrapper/api/v1/server/restapi/operations"
	"github.com/go-openapi/loads"
)

// NewAgentOpenAPIUnixServer instantiates a new instance of the agent OpenAPI server on the unix.
func NewAgentOpenAPIUnixServer() (*IPAMRestapi.Server, error) {
	// read yaml spec
	swaggerSpec, err := loads.Embedded(IPAMRestapi.SwaggerJSON, IPAMRestapi.FlatSwaggerJSON)
	if nil != err {
		return nil, err
	}

	// create new service API
	api := IPAMOperation.NewIpamwrapperAgentAPIAPI(swaggerSpec)

	// daemonset API
	api.IpamwrapperAgentPostIpamHandler = unixPostIp
	api.IpamwrapperAgentDeleteIpamHandler = unixDeleteIp
	// new agent OpenAPI server with api
	srv := IPAMRestapi.NewServer(api)

	// set Unix server with specified unix socket path.
	srv.SocketPath = flags.Filename(ConfigInstance.IpamUnixSocketPath)

	// configure API and handlers with some default values.
	srv.ConfigureAPI()

	return srv, nil
}
