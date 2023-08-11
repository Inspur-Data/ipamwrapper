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
	"fmt"
	"github.com/Inspur-Data/ipamwrapper/pkg/constant"
	ipamip "github.com/Inspur-Data/ipamwrapper/pkg/ip"
	"github.com/Inspur-Data/ipamwrapper/pkg/types"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var ippoollog = logf.Log.WithName("ippool-resource")

func (r *IPPool) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-inspuripam-inspur-com-v1-ippool,mutating=true,failurePolicy=fail,sideEffects=None,groups=inspuripam.inspur.com,resources=ippools,verbs=create;update,versions=v1,name=mippool.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &IPPool{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *IPPool) Default() {
	ippoollog.Info("default", "name", r.Name)

	if err := r.adjustIPPool(); err != nil {
		ippoollog.Error(err, "adjust ippool failed")
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-inspuripam-inspur-com-v1-ippool,mutating=false,failurePolicy=fail,sideEffects=None,groups=inspuripam.inspur.com,resources=ippools,verbs=create;update,versions=v1,name=vippool.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &IPPool{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *IPPool) ValidateCreate() (warnings admission.Warnings, err error) {
	ippoollog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *IPPool) ValidateUpdate(old runtime.Object) (warnings admission.Warnings, err error) {
	ippoollog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *IPPool) ValidateDelete() (warnings admission.Warnings, err error) {
	ippoollog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}

// adjustIPPool check the base info about a ippool
func (r *IPPool) adjustIPPool() error {
	//ippool is deleting
	if r.DeletionTimestamp != nil {
		ippoollog.Info("ippool is deleting, noting to adjust")
		return nil
	}

	//check the finalizer
	if !controllerutil.ContainsFinalizer(r, constant.IPAMFinalizer) {
		controllerutil.AddFinalizer(r, constant.IPAMFinalizer)
		ippoollog.Info("add finalizer")
	}

	//set ip version
	if r.Spec.IPVersion == nil {
		var version types.IPVersion
		if ipamip.IsIPv4CIDR(r.Spec.CIDR) {
			version = constant.IPv4
		} else if ipamip.IsIPv6CIDR(r.Spec.CIDR) {
			version = constant.IPv6
		} else {
			return fmt.Errorf("failed to generate 'spec.ipVersion' from 'spec.subnet' %s, nothing to mutate", r.Spec.CIDR)
		}
		r.Spec.IPVersion = new(types.IPVersion)
		*r.Spec.IPVersion = version
		ippoollog.Info("Set 'spec.ipVersion' to %d", version)
	}

	//merge ips
	if len(r.Spec.IPs) > 1 {
		mergedIPs, err := ipamip.MergeIPRanges(*r.Spec.IPVersion, r.Spec.IPs)
		if err != nil {
			return fmt.Errorf("failed to merge ips: %v", err)
		}

		ips := r.Spec.IPs
		r.Spec.IPs = mergedIPs
		ippoollog.Info("Merge ips: %v to %v", ips, mergedIPs)
	}

	if len(r.Spec.ExcludeIPs) > 1 {
		mergedExcludeIPs, err := ipamip.MergeIPRanges(*r.Spec.IPVersion, r.Spec.ExcludeIPs)
		if err != nil {
			return fmt.Errorf("failed to merge ips: %v", err)
		}

		excludeIPs := r.Spec.ExcludeIPs
		r.Spec.ExcludeIPs = mergedExcludeIPs
		ippoollog.Info("Merge ips: %v to %v", excludeIPs, mergedExcludeIPs)
	}

	return nil
}
