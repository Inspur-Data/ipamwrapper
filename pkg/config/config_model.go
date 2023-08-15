package config

import (
	"github.com/Inspur-Data/ipamwrapper/api/v1/models"
	"github.com/containernetworking/cni/pkg/types"
	"net"
)

type CNIConf struct {
	Name       string    `json:"name"`
	CNIVersion string    `json:"cniVersion"`
	Type       string    `json:"type"`
	IPAM       *IPAMConf `json:"ipam"`
}
type IPAMConf struct {
	Type           string   `json:"type"`
	Routes         []string `json:"routes"`
	LogFile        string   `json:"log_file"`
	LogLevel       string   `json:"log_level"`
	UnixSocketPath string   `json:"unix_socket_path"`
}

var SupportCniVersion = []string{"0.1.0", "0.2.0", "0.3.0", "0.3.1", "0.4.0", "1.0.0"}

type PodArgs struct {
	types.CommonArgs
	IP                         net.IP
	K8S_POD_NAME               types.UnmarshallableString
	K8S_POD_NAMESPACE          types.UnmarshallableString
	K8S_POD_INFRA_CONTAINER_ID types.UnmarshallableString
	K8S_POD_UID                types.UnmarshallableString
}
type RouterNetConf struct {
	types.NetConf
	RouterConfig *models.RouterConfig  `json:"routers"`
}