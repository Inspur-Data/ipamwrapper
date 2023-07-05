// Package endpointmanager
// Copyright 2022 Authors of Inspur
// SPDX-License-Identifier: Apache-2.0
package endpointmanager

import (
	"context"
	"github.com/Inspur-Data/ipamwrapper/pkg/logging"

	"github.com/Inspur-Data/ipamwrapper/pkg/constant"
	inspuripamv1 "github.com/Inspur-Data/ipamwrapper/pkg/k8s/api/v1"
	"github.com/Inspur-Data/ipamwrapper/pkg/types"
	"github.com/Inspur-Data/ipamwrapper/pkg/utils/convert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apitypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type EndpointManager interface {
	GetEndpointByName(ctx context.Context, namespace, podName string, cached bool) (*inspuripamv1.IPAMEndpoint, error)
	ListEndpoints(ctx context.Context, cached bool, opts ...client.ListOption) (*inspuripamv1.IPAMEndpointList, error)
	DeleteEndpoint(ctx context.Context, endpoint *inspuripamv1.IPAMEndpoint) error
	RemoveFinalizer(ctx context.Context, endpoint *inspuripamv1.IPAMEndpoint) error
	PatchIPAllocationResults(ctx context.Context, results []*types.AllocationResult, endpoint *inspuripamv1.IPAMEndpoint, pod *corev1.Pod, podController types.PodTopController) error
	UpdateEndpoint(ctx context.Context, uid, nodeName string, endpoint *inspuripamv1.IPAMEndpoint) error
	ReuseExistIP(uid, nic string, endpoint *inspuripamv1.IPAMEndpoint) *inspuripamv1.IPAMEndpointStatus
	IsValidEndpoint(uid, nic string, endpoint *inspuripamv1.IPAMEndpoint, sts bool) bool
}

type endpointManager struct {
	client    client.Client
	apiReader client.Reader
}

func NewEndpointManager(client client.Client, apiReader client.Reader) (EndpointManager, error) {
	if client == nil {
		return nil, logging.Errorf("api client is nil")
	}
	if apiReader == nil {
		return nil, logging.Errorf("api reader is nil")
	}

	return &endpointManager{
		client:    client,
		apiReader: apiReader,
	}, nil
}

func (em *endpointManager) GetEndpointByName(ctx context.Context, namespace, podName string, cached bool) (*inspuripamv1.IPAMEndpoint, error) {
	reader := em.apiReader
	if cached == constant.UseCache {
		reader = em.client
	}

	var endpoint inspuripamv1.IPAMEndpoint
	if err := reader.Get(ctx, apitypes.NamespacedName{Namespace: namespace, Name: podName}, &endpoint); nil != err {
		return nil, err
	}

	return &endpoint, nil
}

func (em *endpointManager) ListEndpoints(ctx context.Context, cached bool, opts ...client.ListOption) (*inspuripamv1.IPAMEndpointList, error) {
	reader := em.apiReader
	if cached == constant.UseCache {
		reader = em.client
	}

	var endpointList inspuripamv1.IPAMEndpointList
	if err := reader.List(ctx, &endpointList, opts...); err != nil {
		return nil, err
	}

	return &endpointList, nil
}

func (em *endpointManager) DeleteEndpoint(ctx context.Context, endpoint *inspuripamv1.IPAMEndpoint) error {
	if err := em.client.Delete(ctx, endpoint); err != nil {
		return client.IgnoreNotFound(err)
	}

	return nil
}

func (em *endpointManager) RemoveFinalizer(ctx context.Context, endpoint *inspuripamv1.IPAMEndpoint) error {
	if endpoint == nil {
		return logging.Errorf("endpoint is nil")
	}

	if !controllerutil.ContainsFinalizer(endpoint, constant.IPAMFinalizer) {
		return nil
	}

	oldEndpoint := endpoint.DeepCopy()
	controllerutil.RemoveFinalizer(endpoint, constant.IPAMFinalizer)

	if err := em.client.Patch(ctx, endpoint, client.MergeFrom(oldEndpoint)); err != nil {
		return logging.Errorf("failed to remove finalizer %s from endpoint %s/%s: %v", constant.IPAMFinalizer, endpoint.Namespace, endpoint.Name, err)
	}

	return nil
}

func (em *endpointManager) PatchIPAllocationResults(ctx context.Context, results []*types.AllocationResult, endpoint *inspuripamv1.IPAMEndpoint, pod *corev1.Pod, podController types.PodTopController) error {
	if pod == nil {
		return logging.Errorf("pod is nil")
	}

	if endpoint == nil {

		endpoint = &inspuripamv1.IPAMEndpoint{
			ObjectMeta: metav1.ObjectMeta{
				Name:      pod.Name,
				Namespace: pod.Namespace,
			},
			Spec: inspuripamv1.IPAMEndpointSpec{},
		}

		// Do not set ownerReference for Endpoint when its corresponding Pod is
		// controlled by StatefulSet. Once the Pod of StatefulSet is recreated,
		// we can immediately retrieve the old IP allocation results from the
		// Endpoint without worrying about the cascading deletion of the Endpoint.
		if podController.Kind != constant.KindStatefulSet {
			if err := controllerutil.SetOwnerReference(pod, endpoint, em.client.Scheme()); err != nil {
				logging.Errorf("SetOwnerReference failed:%v", err)
				return err
			}
		}
		controllerutil.AddFinalizer(endpoint, constant.IPAMFinalizer)
		err := em.client.Create(ctx, endpoint)
		if err != nil {
			logging.Errorf("create endpoint failed:%v", err)
			return err
		}

		//update the endpoint's status
		endpoint.Status = inspuripamv1.IPAMEndpointStatus{
			IPs:      convert.ConvertResultsToIPDetails(results),
			UID:      string(pod.UID),
			Node:     pod.Spec.NodeName,
			TopOwner: podController.Kind,
		}
		err = em.client.Status().Update(ctx, endpoint)
		if err != nil {
			logging.Errorf("update endpoint failed:%v", err)
			return err
		}
		return nil
	}

	//todo add pod UID

	if endpoint.Status.UID != string(pod.UID) {
		return nil
	}

	// TODO: Only append records with different NIC.
	endpoint.Status.IPs = append(endpoint.Status.IPs, convert.ConvertResultsToIPDetails(results)...)
	return em.client.Status().Update(ctx, endpoint)
}

func (em *endpointManager) UpdateEndpoint(ctx context.Context, uid, nodeName string, endpoint *inspuripamv1.IPAMEndpoint) error {
	if endpoint == nil {
		return logging.Errorf("endpoint is nil")
	}

	if endpoint.Status.UID == uid {
		return nil
	}

	endpoint.Status.UID = uid
	endpoint.Status.Node = nodeName

	return em.client.Status().Update(ctx, endpoint)
}

// GetEndpointIP will return the ips about the endpoint
func GetEndpointIP(uid, nic string, endpoint *inspuripamv1.IPAMEndpoint, isSTS bool) *inspuripamv1.IPAMEndpointStatus {
	if endpoint == nil {
		return nil
	}

	if endpoint.Status.UID == uid || isSTS {
		for _, d := range endpoint.Status.IPs {
			if *d.NIC == nic {
				return &endpoint.Status
			}
		}
	}

	return nil
}

func (em *endpointManager) IsValidEndpoint(uid, nic string, endpoint *inspuripamv1.IPAMEndpoint, sts bool) bool {
	if endpoint == nil {
		return false
	}

	if endpoint.Status.UID == uid || sts {
		for _, d := range endpoint.Status.IPs {
			if *d.NIC == nic {
				return true
			}
		}
	}
	return false
}

// ReuseExistIP will reuse the ip has been allocated
func (em *endpointManager) ReuseExistIP(uid, nic string, endpoint *inspuripamv1.IPAMEndpoint) *inspuripamv1.IPAMEndpointStatus {
	if endpoint == nil {
		return nil
	}

	if endpoint.Status.UID == uid {
		for _, d := range endpoint.Status.IPs {
			if *d.NIC == nic {
				return &endpoint.Status
			}
		}
	}

	return nil
}
