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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var ipamendpointlog = logf.Log.WithName("ipamendpoint-resource")

func (r *IPAMEndpoint) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-inspuripam-inspur-com-v1-ipamendpoint,mutating=true,failurePolicy=fail,sideEffects=None,groups=inspuripam.inspur.com,resources=ipamendpoints,verbs=create;update,versions=v1,name=mipamendpoint.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &IPAMEndpoint{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *IPAMEndpoint) Default() {
	ipamendpointlog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-inspuripam-inspur-com-v1-ipamendpoint,mutating=false,failurePolicy=fail,sideEffects=None,groups=inspuripam.inspur.com,resources=ipamendpoints,verbs=create;update,versions=v1,name=vipamendpoint.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &IPAMEndpoint{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *IPAMEndpoint) ValidateCreate() (warnings admission.Warnings, err error) {
	ipamendpointlog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *IPAMEndpoint) ValidateUpdate(old runtime.Object) (warnings admission.Warnings, err error) {
	ipamendpointlog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *IPAMEndpoint) ValidateDelete() (warnings admission.Warnings, err error) {
	ipamendpointlog.Info("validate delete", "name", r.Name)
	// TODO(user): fill in your validation logic upon object deletion.
	/*
		errList := r.validDelete()
		if len(errList) != 0 {
			return nil, apierrors.NewInvalid(schema.GroupKind{Group: constant.APIGroup, Kind: constant.ENDPOINTS}, r.Name, errList)
		}*/
	return nil, nil
}
func (r *IPAMEndpoint) validDelete() field.ErrorList {
	if len(r.Status.IPs) > 0 {
		err := field.InternalError(allocateIPField, fmt.Errorf("endpoint:%s  still has allocated IPs", r.Name))
		return field.ErrorList{err}
	}
	return nil
}
