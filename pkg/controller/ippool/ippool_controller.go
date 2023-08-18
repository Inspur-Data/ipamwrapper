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

package ippoolcontroller

import (
	"context"
	"github.com/Inspur-Data/ipamwrapper/pkg/constant"
	ipamip "github.com/Inspur-Data/ipamwrapper/pkg/ip"
	"github.com/Inspur-Data/ipamwrapper/pkg/logging"
	"github.com/Inspur-Data/ipamwrapper/pkg/utils/convert"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	inspuripamv1 "github.com/Inspur-Data/ipamwrapper/pkg/k8s/api/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// IPPoolReconciler reconciles a IPPool object
type IPPoolReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=inspuripam.inspur.com,resources=ippools,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=inspuripam.inspur.com,resources=ippools/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=inspuripam.inspur.com,resources=ippools/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the IPPool object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *IPPoolReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here
	ippool := inspuripamv1.IPPool{}
	err := r.Get(ctx, req.NamespacedName, &ippool)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logging.Debugf("ippool %v has been deleted", req.NamespacedName.String())
			return ctrl.Result{}, nil
		} else {
			logging.Errorf("get ippool %v failed, err : %v", req.NamespacedName.String(), err)
			return ctrl.Result{}, err
		}
	}

	//ippool is deleting
	if ippool.DeletionTimestamp != nil {
		logging.Debugf("ippool:%s is deleting...", req.NamespacedName.String())
		var needDel bool
		// ippool has no allocated ip,can direct delete
		if ippool.Status.AllocatedIPs == nil || *(ippool.Status.AllocatedIPCount) == 0 {
			needDel = true
		}
		if needDel {
			err := r.removeFinalizer(ctx, &ippool)
			if nil != err {
				if apierrors.IsNotFound(err) {
					logging.Debugf("ippool:%v has been deleted", req.NamespacedName.String())
					return ctrl.Result{}, nil
				} else {
					logging.Errorf("remove the ippool :%s  finalizer failed: %v", req.NamespacedName.String(), err)
					return ctrl.Result{}, err
				}
			} else {
				logging.Debugf("remove ippool: '%s' finalizer successfully", req.NamespacedName.String())
			}
		}
	} else {
		// ippool is created or updated
		logging.Debugf("ippool:%s is creating or updating ...", req.NamespacedName.String())
		needUpdate := false
		// initial the original data
		if ippool.Status.AllocatedIPCount == nil {
			needUpdate = true
			ippool.Status.AllocatedIPCount = pointer.Int64(0)
			logging.Debugf("set ippool  '%s' status AllocatedIPCount to 0", req.NamespacedName.String())
		}

		allocatedIPs, err := convert.UnmarshalIPPoolAllocatedIPs(ippool.Status.AllocatedIPs)
		if nil != err {

		}

		if int64(len(allocatedIPs)) != *ippool.Status.AllocatedIPCount {
			needUpdate = true
			ippool.Status.AllocatedIPCount = pointer.Int64(int64(len(allocatedIPs)))
			logging.Debugf("allocateIPCount unequal to length of the allocatedIPs, set ippool  '%s' status AllocatedIPCount to %d", req.NamespacedName, len(allocatedIPs))
		}

		totalIPs, err := ipamip.AssembleTotalIPs(*ippool.Spec.IPVersion, ippool.Spec.IPs, ippool.Spec.ExcludeIPs)
		if nil != err {
			logging.Errorf("calculate total ip failed: %v", err)
			return ctrl.Result{}, err
		}

		if ippool.Status.TotalIPCount == nil || *ippool.Status.TotalIPCount != int64(len(totalIPs)) {
			needUpdate = true
			ippool.Status.TotalIPCount = pointer.Int64(int64(len(totalIPs)))
		}

		if needUpdate {
			err = r.Client.Status().Update(ctx, &ippool)
			if nil != err {
				logging.Errorf("update ippool failed :%v", err)
				return ctrl.Result{}, err
			}
			logging.Debugf("update ippool '%s' status  successfully", req.NamespacedName.String())
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *IPPoolReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&inspuripamv1.IPPool{}).
		Complete(r)
}

// removeFinalizer removes IPpool's  finalizer
func (r *IPPoolReconciler) removeFinalizer(ctx context.Context, ippool *inspuripamv1.IPPool) error {
	if !controllerutil.ContainsFinalizer(ippool, constant.IPAMFinalizer) {
		return nil
	}

	controllerutil.RemoveFinalizer(ippool, constant.IPAMFinalizer)
	err := r.Client.Update(ctx, ippool)
	if nil != err {
		return err
	}

	return nil
}
