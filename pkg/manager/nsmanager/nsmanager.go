// package nsmanager
// Copyright 2022 Authors of Inspur
// SPDX-License-Identifier: Apache-2.0
package nsmanager

import (
	"context"
	"encoding/json"
	"github.com/Inspur-Data/ipamwrapper/pkg/logging"
	"github.com/Inspur-Data/ipamwrapper/pkg/types"

	corev1 "k8s.io/api/core/v1"
	apitypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Inspur-Data/ipamwrapper/pkg/constant"
)

type NsManager interface {
	GetNamespace(ctx context.Context, nsName string, cached bool) (*corev1.Namespace, error)
	ListNamespaces(ctx context.Context, cached bool, opts ...client.ListOption) (*corev1.NamespaceList, error)
	GetNSDefaultPools(ns *corev1.Namespace) ([]string, []string, error)
}

type nsManager struct {
	client    client.Client
	apiReader client.Reader
}

// NewNamespaceManager will return ns manager instance
func NewNamespaceManager(client client.Client, apiReader client.Reader) (NsManager, error) {
	if client == nil {
		return nil, logging.Errorf("k8s client is nil")
	}
	if apiReader == nil {
		return nil, logging.Errorf("k8s apireader is nil")
	}

	return &nsManager{
		client:    client,
		apiReader: apiReader,
	}, nil
}

// GetNamespace return the ns by name
func (nm *nsManager) GetNamespace(ctx context.Context, nsName string, cached bool) (*corev1.Namespace, error) {
	reader := nm.apiReader
	if cached == constant.UseCache {
		reader = nm.client
	}

	var ns corev1.Namespace
	if err := reader.Get(ctx, apitypes.NamespacedName{Name: nsName}, &ns); err != nil {
		return nil, err
	}

	return &ns, nil
}

// ListNamespaces return the list of the ns
func (nm *nsManager) ListNamespaces(ctx context.Context, cached bool, opts ...client.ListOption) (*corev1.NamespaceList, error) {
	reader := nm.apiReader
	if cached == constant.UseCache {
		reader = nm.client
	}

	var nsList corev1.NamespaceList
	if err := reader.List(ctx, &nsList, opts...); err != nil {
		return nil, err
	}

	return &nsList, nil
}

func (nm *nsManager) GetNSDefaultPools(ns *corev1.Namespace) ([]string, []string, error) {
	if ns == nil {
		return nil, nil, logging.Errorf("namespace is nil")
	}

	var nsDefaultV4Pool types.AnnoNSDefautlV4PoolValue
	var nsDefaultV6Pool types.AnnoNSDefautlV6PoolValue
	var annoPodIPPool types.AnnoPodIPPoolValue
	if v, ok := ns.Annotations[constant.AnnoNSDefautlPool]; ok {
		if err := json.Unmarshal([]byte(v), &annoPodIPPool); err != nil {
			return nil, nil, err
		}
	}

	if annoPodIPPool.IPv4Pools != nil {
		nsDefaultV4Pool = annoPodIPPool.IPv4Pools
	}
	if annoPodIPPool.IPv6Pools != nil {
		nsDefaultV6Pool = annoPodIPPool.IPv4Pools
	}

	return nsDefaultV4Pool, nsDefaultV6Pool, nil
}
