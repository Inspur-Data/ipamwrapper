package daemonset

import (
	"context"
	"github.com/Inspur-Data/ipamwrapper/pkg/logging"
	"net"
	"net/http"

	runtimeClient "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"

	openAPIClient "github.com/Inspur-Data/ipamwrapper/api/v1/client"
)

// NewAgentOpenAPIUnixClient creates a new instance of the agent OpenAPI unix client.
func NewAgentOpenAPIHttpClient(host string) (*openAPIClient.IpamwrapperAgentAPI, error) {
	if host == "" {
		return nil, logging.Errorf("host is nil")
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			DisableCompression: true,
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("tcp", host)
			},
			DisableKeepAlives: true,
		},
	}
	clientTrans := runtimeClient.NewWithClient(host, openAPIClient.DefaultBasePath,
		openAPIClient.DefaultSchemes, httpClient)
	client := openAPIClient.New(clientTrans, strfmt.Default)
	return client, nil
}
