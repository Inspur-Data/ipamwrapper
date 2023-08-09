/*
Copyright 2023 Inspur-Data.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// IPPoolSpec defines the desired state of IPPool
type IPPoolSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +kubebuilder:validation:Enum=4;6
	// +kubebuilder:validation:Optional
	IPVersion *int64 `json:"ipVersion,omitempty"`

	// +kubebuilder:validation:Required
	CIDR string `json:"cidr"`

	// +kubebuilder:validation:Optional
	IPs []string `json:"ips,omitempty"`

	// +kubebuilder:validation:Optional
	ExcludeIPs []string `json:"excludeIPs,omitempty"`

	// +kubebuilder:validation:Optional
	Gateway *string `json:"gateway,omitempty"`

	// +kubebuilder:validation:Optional
	Routes []*Route `json:"routes,omitempty"`

	// +kubebuilder:validation:Optional
	NamespaceAffinity *metav1.LabelSelector `json:"namespaceAffinity,omitempty"`

	// +kubebuilder:validation:Optional
	NodeAffinity *metav1.LabelSelector `json:"nodeAffinity,omitempty"`

	// +kubebuilder:default=false
	// +kubebuilder:validation:Optional
	Default *bool `json:"default,omitempty"`

	// +kubebuilder:default=false
	// +kubebuilder:validation:Optional
	Disable *bool `json:"disable,omitempty"`
}
type Route struct {
	// +kubebuilder:validation:Required
	Dst string `json:"dst"`

	// +kubebuilder:validation:Required
	Gw string `json:"gw"`
}

// PoolIPAllocations is a map of IP allocation details indexed by IP address.
type PoolIPAllocations map[string]PoolIPAllocation

type PoolIPAllocation struct {
	NIC            string `json:"interface"`
	NamespacedName string `json:"pod"`
	PodUID         string `json:"podUid"`
}

// IPPoolStatus defines the observed state of IPPool
type IPPoolStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// +kubebuilder:validation:Optional
	AllocatedIPs *string `json:"allocatedIPs,omitempty"`

	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Optional
	TotalIPCount *int64 `json:"totalIPCount,omitempty"`

	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Optional
	AllocatedIPCount *int64 `json:"allocatedIPCount,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".spec.ipVersion",description="ipVersion",name="IPVERSION",type=string
// +kubebuilder:printcolumn:JSONPath=".spec.cidr",description="cidr",name="CIDR",type=string
// +kubebuilder:printcolumn:JSONPath=".status.allocatedIPCount",description="allocatedIPCount",name="ALLOCATED",type=integer
// +kubebuilder:printcolumn:JSONPath=".status.totalIPCount",description="totalIPCount",name="TOTAL",type=integer
// IPPool is the Schema for the ippools API
type IPPool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IPPoolSpec   `json:"spec,omitempty"`
	Status IPPoolStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// IPPoolList contains a list of IPPool
type IPPoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IPPool `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IPPool{}, &IPPoolList{})
}
