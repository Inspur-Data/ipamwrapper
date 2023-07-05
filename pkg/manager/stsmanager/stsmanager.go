// Package stsmanager
// Copyright 2022 Authors of Inspur
// SPDX-License-Identifier: Apache-2.0

package stsmanager

import (
	"context"
	"fmt"
	"github.com/Inspur-Data/ipamwrapper/pkg/logging"
	"regexp"
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	apitypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Inspur-Data/ipamwrapper/pkg/constant"
)

type StatefulSetManager interface {
	GetStsByName(ctx context.Context, namespace, name string, cached bool) (*appsv1.StatefulSet, error)
	ListSts(ctx context.Context, cached bool, opts ...client.ListOption) (*appsv1.StatefulSetList, error)
	IsValidStsPod(ctx context.Context, namespace, podName, podTopOwner string) (bool, error)
}

type statefulSetManager struct {
	client    client.Client
	apiReader client.Reader
}

// statefulPodRegex is a regular expression that extracts the parent sts and ordinal from the name of a pod
var statefulPodRegex = regexp.MustCompile("(.*)-([0-9]+)$")

// NewStatefulSetManager return a stsmanager client
func NewStatefulSetManager(client client.Client, apiReader client.Reader) (StatefulSetManager, error) {
	if client == nil {
		return nil, fmt.Errorf("k8s client is nil")
	}
	if apiReader == nil {
		return nil, fmt.Errorf("api reader is nil")
	}

	return &statefulSetManager{
		client:    client,
		apiReader: apiReader,
	}, nil
}

// GetStsByName return a statefulset instance by the name
func (sm *statefulSetManager) GetStsByName(ctx context.Context, namespace, name string, cached bool) (*appsv1.StatefulSet, error) {
	reader := sm.apiReader
	if cached == constant.UseCache {
		reader = sm.client
	}

	var sts appsv1.StatefulSet
	if err := reader.Get(ctx, apitypes.NamespacedName{Namespace: namespace, Name: name}, &sts); err != nil {
		logging.Errorf("get sts failed:%v", err)
		return nil, err
	}

	return &sts, nil
}

// ListSts return a statefulset list on the cluster
func (sm *statefulSetManager) ListSts(ctx context.Context, cached bool, opts ...client.ListOption) (*appsv1.StatefulSetList, error) {
	reader := sm.apiReader
	if cached == constant.UseCache {
		reader = sm.client
	}

	var stsList appsv1.StatefulSetList
	if err := reader.List(ctx, &stsList, opts...); err != nil {
		logging.Errorf("list sts fialed :%v", err)
		return nil, err
	}

	return &stsList, nil
}

// IsValidStsPod  will check the pod whether need to be cleaned up with the params
// once the pod's controller StatefulSet was deleted, the pod's corresponding IPPool IP and Endpoint need to be cleaned up.
// or the pod's controller StatefulSet decreased its replicas and the pod's index is out of replicas, it needs to be cleaned up too.
func (sm *statefulSetManager) IsValidStsPod(ctx context.Context, namespace, podName, podControllerType string) (bool, error) {
	if podControllerType != constant.KindStatefulSet {
		return false, fmt.Errorf("pod '%s/%s' is controlled by '%s' instead of StatefulSet", namespace, podName, podControllerType)
	}

	stsName, replicas, found := getStatefulSetNameAndOrdinal(podName)
	if !found {
		return false, nil
	}

	sts, err := sm.GetStsByName(ctx, namespace, stsName, constant.IgnoreCache)
	if err != nil {
		logging.Errorf("get stateful set :%s failed :% +v", stsName, err)
		return false, client.IgnoreNotFound(err)
	}

	// the pod controlled by StatefulSet is created or re-created.
	if replicas <= int(*sts.Spec.Replicas)-1 {
		return true, nil
	}

	return false, nil
}

// getStatefulSetNameAndOrdinal gets the name of pod's parent StatefulSet and pod's ordinal as extracted from its Name. If
// the Pod was not created by a StatefulSet, its parent is considered to be empty string, and its ordinal is considered
// to be -1.
func getStatefulSetNameAndOrdinal(podName string) (parent string, ordinal int, found bool) {
	parent = ""
	ordinal = -1

	subMatches := statefulPodRegex.FindStringSubmatch(podName)
	if len(subMatches) < 3 {
		return parent, ordinal, false
	}

	parent = subMatches[1]
	i, err := strconv.ParseInt(subMatches[2], 10, 32)
	if err != nil {
		return parent, ordinal, false
	}

	ordinal = int(i)
	return parent, ordinal, true
}
