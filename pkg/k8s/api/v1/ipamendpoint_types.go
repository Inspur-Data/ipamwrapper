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

type IPAllocationDetail struct {
	// +kubebuilder:validation:Optional
	NIC *string `json:"interface,omitempty"`

	// +kubebuilder:validation:Optional
	IPv4 *string `json:"ipv4,omitempty"`

	// +kubebuilder:validation:Optional
	IPv6 *string `json:"ipv6,omitempty"`

	// +kubebuilder:validation:Optional
	IPv4Pool *string `json:"ipv4Pool,omitempty"`

	// +kubebuilder:validation:Optional
	IPv6Pool *string `json:"ipv6Pool,omitempty"`

	// +kubebuilder:validation:Optional
	IPv4Gateway *string `json:"ipv4Gateway,omitempty"`

	// +kubebuilder:validation:Optional
	IPv6Gateway *string `json:"ipv6Gateway,omitempty"`

	// +kubebuilder:validation:Optional
	CleanGateway *bool `json:"cleanGateway,omitempty"`

	// +kubebuilder:validation:Optional
	Routes []*Route `json:"routes,omitempty"`
}

// IPAMEndpointSpec defines the desired state of IPAMEndpoint
type IPAMEndpointSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

}

// IPAMEndpointStatus defines the observed state of IPAMEndpoint
type IPAMEndpointStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// +kubebuilder:validation:Optional
	IPs []IPAllocationDetail `json:"ips,omitempty"`

	// +kubebuilder:validation:Optional
	UID string `json:"uid,omitempty"`

	// +kubebuilder:validation:Optional
	Node string `json:"node,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// IPAMEndpoint is the Schema for the ipamendpoints API
type IPAMEndpoint struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec IPAMEndpointSpec `json:"spec,omitempty"`

	Status IPAMEndpointStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// IPAMEndpointList contains a list of IPAMEndpoint
type IPAMEndpointList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IPAMEndpoint `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IPAMEndpoint{}, &IPAMEndpointList{})
}
