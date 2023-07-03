package ipam

import "time"

type IPAMConfig struct {
	EnableIPv4               bool
	EnableIPv6               bool
	ClusterDefaultIPv4IPPool []string
	ClusterDefaultIPv6IPPool []string

	EnableSpiderSubnet bool
	EnableStatefulSet  bool

	OperationRetries     int
	OperationGapDuration time.Duration
	IPv4ReservedIP       []string
	IPv6ReservedIP       []string
}
