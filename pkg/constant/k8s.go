// Copyright 2022 Authors of Inspur
// SPDX-License-Identifier: Apache-2.0

package constant

import (
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

var K8sKinds = []string{KindPod, KindDeployment, KindReplicaSet, KindDaemonSet, KindStatefulSet, KindJob, KindCronJob}
var K8sAPIVersions = []string{corev1.SchemeGroupVersion.String(), appsv1.SchemeGroupVersion.String(), batchv1.SchemeGroupVersion.String()}

const (
	UseCache    = true
	IgnoreCache = false
)

const (
	True  = "true"
	False = "false"
)

const ClusterDefaultInterfaceName = "eth0"
