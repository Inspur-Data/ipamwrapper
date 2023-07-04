// Package ippoolmanager
// Copyright 2022 Authors of Inspur
// SPDX-License-Identifier: Apache-2.0
package ippoolmanager

import (
	"context"
	"k8s.io/client-go/tools/cache"

	ipamip "github.com/Inspur-Data/ipamwrapper/pkg/ip"
	"github.com/Inspur-Data/ipamwrapper/pkg/logging"
	"github.com/Inspur-Data/ipamwrapper/pkg/utils/convert"
	"github.com/Inspur-Data/ipamwrapper/pkg/utils/retry"
	"net"

	"github.com/Inspur-Data/ipamwrapper/api/v1/models"
	"github.com/Inspur-Data/ipamwrapper/pkg/constant"
	inspuripamv1 "github.com/Inspur-Data/ipamwrapper/pkg/k8s/api/v1"
	"github.com/Inspur-Data/ipamwrapper/pkg/types"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apitypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const defaultMaxAllocatedIPs = 1024

type IPPoolManager interface {
	GetIPPoolByName(ctx context.Context, poolName string, cached bool) (*inspuripamv1.IPPool, error)
	ListIPPools(ctx context.Context, cached bool, opts ...client.ListOption) (*inspuripamv1.IPPoolList, error)
	AllocateIP(ctx context.Context, pool *inspuripamv1.IPPool, nic string, pod *corev1.Pod, ipv4ReservedIPs []string, ipv6ReservedIPs []string) (*models.IPConfig, error)
	ReleaseIP(ctx context.Context, poolName string, ipAndUIDs []types.IPAndUID) error
	UpdateAllocatedIPs(ctx context.Context, poolName string, ipAndCIDs []types.IPAndUID) error
	UpdateIPPoolStatus(ctx context.Context, ipPool *inspuripamv1.IPPool) error
}

type ipPoolManager struct {
	config    MgrConfig
	client    client.Client
	apiReader client.Reader
	//todo reserved IP
	//rIPManager reservedipmanager.ReservedIPManager
}

type MgrConfig struct {
	MaxAllocatedIPs *int
}

func NewIPPoolManager(config MgrConfig, client client.Client, apiReader client.Reader) (IPPoolManager, error) {
	if client == nil {
		return nil, logging.Errorf("k8s client is nil")
	}
	if apiReader == nil {
		return nil, logging.Errorf("api Reader is nil")
	}
	//todo reservedIP check
	/*
		if rIPManager == nil {
			return nil, fmt.Errorf("reserved-IP manager %w", constant.ErrMissingRequiredParam)
		}*/

	return &ipPoolManager{
		config:    setDefaultsForIPPoolManagerConfig(config),
		client:    client,
		apiReader: apiReader,
		//rIPManager: rIPManager,
	}, nil
}

func (im *ipPoolManager) GetIPPoolByName(ctx context.Context, poolName string, cached bool) (*inspuripamv1.IPPool, error) {
	reader := im.apiReader
	if cached == constant.UseCache {
		reader = im.client
	}

	var ipPool inspuripamv1.IPPool
	if err := reader.Get(ctx, apitypes.NamespacedName{Name: poolName, Namespace: "default"}, &ipPool); err != nil {
		return nil, err
	}

	return &ipPool, nil
}

func (im *ipPoolManager) ListIPPools(ctx context.Context, cached bool, opts ...client.ListOption) (*inspuripamv1.IPPoolList, error) {
	reader := im.apiReader
	if cached == constant.UseCache {
		reader = im.client
	}

	var ipPoolList inspuripamv1.IPPoolList
	if err := reader.List(ctx, &ipPoolList, opts...); err != nil {
		return nil, err
	}

	return &ipPoolList, nil
}

func (im *ipPoolManager) AllocateIP(ctx context.Context, ipPool *inspuripamv1.IPPool, nic string, pod *corev1.Pod, ipv4ReservedIPs []string, ipv6ReservedIPs []string) (*models.IPConfig, error) {

	backoff := retry.DefaultRetry
	//steps := backoff.Steps
	var ipConfig *models.IPConfig
	err := retry.RetryOnConflictWithContext(ctx, backoff, func(ctx context.Context) error {
		logging.Debugf("generate a random IP address")
		allocatedIP, err := im.genRandomIP(ctx, nic, ipPool, pod, ipv4ReservedIPs, ipv6ReservedIPs)
		if err != nil {
			return err
		}

		logging.Debugf(" update the allocation status of ippool using random IP %s", allocatedIP)
		if err := im.client.Status().Update(ctx, ipPool); err != nil {
			if apierrors.IsConflict(err) {
				//todo add metrics
				logging.Debugf("update the status of ippool conflict")
			}
			return err
		}
		ipConfig = convert.GenIPConfigResult(allocatedIP, nic, ipPool)

		return nil
	})
	if err != nil {

		return nil, err
	}

	return ipConfig, nil
}

func (im *ipPoolManager) genRandomIP(ctx context.Context, nic string, ipPool *inspuripamv1.IPPool, pod *corev1.Pod, ipv4ReservedIPs []string, ipv6ReservedIPs []string) (net.IP, error) {
	var reservedIPs []net.IP
	var err error
	if *ipPool.Spec.IPVersion == constant.IPv4 {
		reservedIPs, err = ipamip.ParseIPRanges(constant.IPv4, ipv4ReservedIPs)
		if err != nil {
			logging.Errorf("parse ip range failed:%v", err)
		}
	}
	if *ipPool.Spec.IPVersion == constant.IPv6 {
		reservedIPs, err = ipamip.ParseIPRanges(constant.IPv6, ipv6ReservedIPs)
		if err != nil {
			logging.Errorf("parse ip range failed:%v", err)
		}
	}
	allocatedRecords, err := convert.UnmarshalIPPoolAllocatedIPs(ipPool.Status.AllocatedIPs)
	if err != nil {
		return nil, err
	}

	var used []string
	for ip := range allocatedRecords {
		used = append(used, ip)
	}
	usedIPs, err := ipamip.ParseIPRanges(*ipPool.Spec.IPVersion, used)
	if err != nil {
		return nil, err
	}

	totalIPs, err := ipamip.AssembleTotalIPs(*ipPool.Spec.IPVersion, ipPool.Spec.IPs, ipPool.Spec.ExcludeIPs)
	if err != nil {
		return nil, err
	}

	availableIPs := ipamip.IPsDiffSet(totalIPs, append(reservedIPs, usedIPs...), false)
	if len(availableIPs) == 0 {
		return nil, logging.Errorf("ip has been used up in ippool:%s", ipPool.Name)
	}
	resIP := availableIPs[0]

	key, err := cache.MetaNamespaceKeyFunc(pod)
	if err != nil {
		return nil, err
	}

	if allocatedRecords == nil {
		allocatedRecords = inspuripamv1.PoolIPAllocations{}
	}
	allocatedRecords[resIP.String()] = inspuripamv1.PoolIPAllocation{
		NIC:            nic,
		NamespacedName: key,
		PodUID:         string(pod.UID),
	}

	data, err := convert.MarshalIPPoolAllocatedIPs(allocatedRecords)
	if err != nil {
		return nil, err
	}
	ipPool.Status.AllocatedIPs = data

	if ipPool.Status.AllocatedIPCount == nil {
		ipPool.Status.AllocatedIPCount = new(int64)
	}

	*ipPool.Status.AllocatedIPCount++
	if *ipPool.Status.AllocatedIPCount > int64(*im.config.MaxAllocatedIPs) {
		return nil, logging.Errorf("ippool %s ip has exceeded", ipPool.Name)
	}

	return resIP, nil
}

func (im *ipPoolManager) ReleaseIP(ctx context.Context, poolName string, ipAndUIDs []types.IPAndUID) error {

	backoff := retry.DefaultRetry
	//steps := backoff.Steps
	err := retry.RetryOnConflictWithContext(ctx, backoff, func(ctx context.Context) error {
		logging.Debugf(" IPPool for IP release")
		ipPool, err := im.GetIPPoolByName(ctx, poolName, constant.IgnoreCache)
		if err != nil {
			return err
		}

		allocatedRecords, err := convert.UnmarshalIPPoolAllocatedIPs(ipPool.Status.AllocatedIPs)
		if err != nil {
			return err
		}

		if ipPool.Status.AllocatedIPCount == nil {
			ipPool.Status.AllocatedIPCount = new(int64)
		}

		release := false
		for _, iu := range ipAndUIDs {
			if record, ok := allocatedRecords[iu.IP]; ok {
				if record.PodUID == iu.UID {
					delete(allocatedRecords, iu.IP)
					*ipPool.Status.AllocatedIPCount--
					release = true
				}
			}
		}

		if !release {
			return nil
		}

		data, err := convert.MarshalIPPoolAllocatedIPs(allocatedRecords)
		if err != nil {
			return err
		}
		ipPool.Status.AllocatedIPs = data

		if err := im.client.Status().Update(ctx, ipPool); err != nil {
			if apierrors.IsConflict(err) {
				//todo add metrics
				//metric.IpamReleaseUpdateIPPoolConflictCounts.Add(ctx, 1)
				//logger.Debug("An conflict occurred when cleaning the IP allocation records of IPPool")
			}
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
func (im *ipPoolManager) UpdateIPPoolStatus(ctx context.Context, ipPool *inspuripamv1.IPPool) error {
	if err := im.client.Status().Update(ctx, ipPool); err != nil {
		if apierrors.IsConflict(err) {
			//todo add metircs
			//metric.IpamAllocationUpdateIPPoolConflictCounts.Add(ctx, 1)
			logging.Errorf("update ippool failed:%v", err)
		}
		return err
	}

	return nil
}
func (im *ipPoolManager) UpdateAllocatedIPs(ctx context.Context, poolName string, ipAndUIDs []types.IPAndUID) error {
	backoff := retry.DefaultRetry
	//steps := backoff.Steps
	err := retry.RetryOnConflictWithContext(ctx, backoff, func(ctx context.Context) error {
		ipPool, err := im.GetIPPoolByName(ctx, poolName, constant.IgnoreCache)
		if err != nil {
			return err
		}

		allocatedRecords, err := convert.UnmarshalIPPoolAllocatedIPs(ipPool.Status.AllocatedIPs)
		if err != nil {
			return err
		}

		recreate := false
		for _, iu := range ipAndUIDs {
			if record, ok := allocatedRecords[iu.IP]; ok {
				if record.PodUID != iu.UID {
					record.PodUID = iu.UID
					allocatedRecords[iu.IP] = record
					recreate = true
				}
			}
		}

		if !recreate {
			return nil
		}

		data, err := convert.MarshalIPPoolAllocatedIPs(allocatedRecords)
		if err != nil {
			return err
		}
		ipPool.Status.AllocatedIPs = data

		if err := im.client.Status().Update(ctx, ipPool); err != nil {
			if apierrors.IsConflict(err) {
				//todo add metircs
				//metric.IpamAllocationUpdateIPPoolConflictCounts.Add(ctx, 1)
			}
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func setDefaultsForIPPoolManagerConfig(config MgrConfig) MgrConfig {
	if config.MaxAllocatedIPs == nil || *config.MaxAllocatedIPs == 0 {
		maxAllocatedIPs := defaultMaxAllocatedIPs
		config.MaxAllocatedIPs = &maxAllocatedIPs
	}

	return config
}
