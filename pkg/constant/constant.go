// Copyright 2022 Authors of Inspur
// SPDX-License-Identifier: Apache-2.0
package constant

type IPVersion = int64

const (
	IPv4 IPVersion = 4
	IPv6 IPVersion = 6
)

// Network configurations
const (
	NetworkLegacy             = "legacy"
	NetworkStrict             = "strict"
	NetworkSDN                = "sdn"
	DefaultIPAMUnixSocketPath = "/var/run/inspur/ipamwrapper.sock"
)

const (
	KindPod         = "Pod"
	KindDeployment  = "Deployment"
	KindStatefulSet = "StatefulSet"
	KindDaemonSet   = "DaemonSet"
	KindUnknown     = "Unknown"
	KindReplicaSet  = "ReplicaSet"
	KindJob         = "Job"
	KindCronJob     = "CronJob"
)

const (
	PodRunning     = "Running"
	PodTerminating = "Terminating"
	PodSucceeded   = "Succeeded"
	PodFailed      = "Failed"
	PodEvicted     = "Evicted"
	PodDeleted     = "Deleted"
	PodUnknown     = "Unknown"
)

const (
	IPAMFinalizer = "inspur.io"
)

const (
	Pre                 = "ipam.inspur.io"
	AnnoPodIPPool       = Pre + "/ippool"
	AnnoPodIPPools      = Pre + "/ippools"
	AnnoPodRoutes       = Pre + "/routes"
	AnnoPodDNS          = Pre + "/dns"
	AnnoNSDefautlV4Pool = Pre + "/default-ipv4-ippool"
	AnnoNSDefautlV6Pool = Pre + "/default-ipv6-ippool"
	AnnoNSDefautlPool   = Pre + "/default-ippool"
)
