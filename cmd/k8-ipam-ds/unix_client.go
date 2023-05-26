package k8_ipam_ds

import (
	"context"
	"fmt"
	"net"
	"net/http"

	runtime_client "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"

	openAPIClient "github.com/Inspur-Data/k8-ipam/api/v1/client"
)

// NewAgentOpenAPIUnixClient creates a new instance of the agent OpenAPI unix client.
func NewAgentOpenAPIUnixClient(unixSocketPath string) (*openAPIClient.K8IpamAgentAPI, error) {
	if unixSocketPath == "" {
		return nil, fmt.Errorf("unix socket path must be specified")
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			DisableCompression: true,
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", unixSocketPath)
			},
			DisableKeepAlives: true,
		},
	}
	clientTrans := runtime_client.NewWithClient(unixSocketPath, openAPIClient.DefaultBasePath,
		openAPIClient.DefaultSchemes, httpClient)
	client := openAPIClient.New(clientTrans, strfmt.Default)
	return client, nil
}
