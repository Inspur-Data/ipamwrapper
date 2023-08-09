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

package endpointcontroller

import (
	"context"
	"github.com/Inspur-Data/ipamwrapper/pkg/constant"
	"github.com/Inspur-Data/ipamwrapper/pkg/logging"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	inspuripamv1 "github.com/Inspur-Data/ipamwrapper/pkg/k8s/api/v1"
)

// IPAMEndpointReconciler reconciles a IPAMEndpoint object
type IPAMEndpointReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=inspuripam.inspur.com,resources=ipamendpoints,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=inspuripam.inspur.com,resources=ipamendpoints/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=inspuripam.inspur.com,resources=ipamendpoints/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the IPAMEndpoint object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *IPAMEndpointReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here
	endpoint := inspuripamv1.IPAMEndpoint{}
	err := r.Get(ctx, req.NamespacedName, &endpoint)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logging.Debugf("endpoint %v has been deleted", req.NamespacedName.String())
			return ctrl.Result{}, nil
		} else {
			logging.Errorf("get endpoint %v failed, err : %v", req.NamespacedName.String(), err)
			return ctrl.Result{}, err
		}
	}

	if endpoint.DeletionTimestamp != nil {
		logging.Debugf("endpoint :%s is deleting...", req.NamespacedName.String())
		err := r.removeFinalizer(ctx, &endpoint)
		if nil != err {
			if apierrors.IsNotFound(err) {
				logging.Debugf("endpoint:%v has been deleted", req.NamespacedName.String())
				return ctrl.Result{}, nil
			} else {
				logging.Errorf("remove the endpoint :%s  finalizer failed: %v", req.NamespacedName.String(), err)
				return ctrl.Result{}, err
			}
		} else {
			logging.Debugf("remove ippool: '%s' finalizer successfully", req.NamespacedName.String())
		}
	} else {
		logging.Debugf("endpoint :%s is creating or updating...", req.NamespacedName.String())
		//todo add some logic
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *IPAMEndpointReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&inspuripamv1.IPAMEndpoint{}).
		Complete(r)
}

// removeFinalizer removes ipam endpoint's  finalizer
func (r *IPAMEndpointReconciler) removeFinalizer(ctx context.Context, endpoint *inspuripamv1.IPAMEndpoint) error {
	if !controllerutil.ContainsFinalizer(endpoint, constant.IPAMFinalizer) {
		return nil
	}

	controllerutil.RemoveFinalizer(endpoint, constant.IPAMFinalizer)
	err := r.Client.Update(ctx, endpoint)
	if nil != err {
		return err
	}

	return nil
}
