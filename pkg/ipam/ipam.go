// Copyright 2022 Authors of Inspur
// SPDX-License-Identifier: Apache-2.0

package ipam

import (
	"context"
	"github.com/Inspur-Data/ipamwrapper/api/v1/models"
	"github.com/Inspur-Data/ipamwrapper/pkg/constant"
	"github.com/Inspur-Data/ipamwrapper/pkg/logging"
	"github.com/Inspur-Data/ipamwrapper/pkg/manager/endpointmanager"
	"github.com/Inspur-Data/ipamwrapper/pkg/manager/ippoolmanager"
	"github.com/Inspur-Data/ipamwrapper/pkg/manager/nsmanager"
	"github.com/Inspur-Data/ipamwrapper/pkg/manager/podmanager"
	"github.com/Inspur-Data/ipamwrapper/pkg/manager/stsmanager"
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

type IPAM interface {
	Allocate(ctx context.Context, addArgs *models.IpamAllocArgs) (*models.IpamAllocResponse, error)
	Delete(ctx context.Context, delArgs *models.IpamDelArgs) error
	Start(ctx context.Context) error
}

type ipam struct {
	config          IPAMConfig
	podManager      podmanager.PodManager
	endpointManager endpointmanager.EndpointManager
	ippoolManager   ippoolmanager.IPPoolManager
	nsManager       nsmanager.NsManager
	stsManager      stsmanager.StatefulSetManager
}

// NewIPAM init a new IPAM instance
func NewIPAM(config IPAMConfig, podMgr podmanager.PodManager,
	endpointMgr endpointmanager.EndpointManager,
	ippoolMgr ippoolmanager.IPPoolManager,
	nsMgr nsmanager.NsManager,
	stsMgr stsmanager.StatefulSetManager) (IPAM, error) {
	if podMgr == nil {
		return nil, logging.Errorf("podManager is nil")
	}

	if endpointMgr == nil {
		return nil, logging.Errorf("endpointManager is nil")
	}

	if ippoolMgr == nil {
		return nil, logging.Errorf("ippoolManager is nil")
	}

	if nsMgr == nil {
		return nil, logging.Errorf("nsManager is nil")
	}

	if stsMgr == nil {
		return nil, logging.Errorf("stsManager is nil")
	}

	return &ipam{
		podManager:      podMgr,
		config:          config,
		endpointManager: endpointMgr,
		ippoolManager:   ippoolMgr,
		nsManager:       nsMgr,
		stsManager:      stsMgr,
	}, nil
}

// Allocate will allocate an IP with the given param
func (i *ipam) Allocate(ctx context.Context, addArgs *models.IpamAllocArgs) (*models.IpamAllocResponse, error) {
	pod, err := i.podManager.GetPodByName(ctx, *addArgs.PodNamespace, *addArgs.PodName, true)
	if err != nil {
		logging.Errorf("get pod failed:%v", err)
		return nil, err
	}

	//get pod's top owner,if the top owner is sts,it will return directly
	owner, err := i.podManager.GetPodTopOwner(ctx, pod)
	if err != nil {
		logging.Errorf("get pod top owner failed:%v", err)
		return nil, err
	}

	//get endpoint
	ed, err := i.endpointManager.GetEndpointByName(ctx, *addArgs.PodNamespace, *addArgs.PodName, true)
	if client.IgnoreNotFound(err) != nil {
		logging.Errorf("get endpoint failed:%v", err)
		return nil, err
	}
	if ed != nil {
		logging.Debugf("get endpoint %s/%s", pod.Namespace, pod.Name)
	} else {
		logging.Errorf("find no endpoint")
	}

	if i.config.EnableStatefulSet && owner.APIVersion == appsv1.SchemeGroupVersion.String() && owner.Kind == constant.KindStatefulSet {
		logging.Debugf("owner is statefulset,try to reuse the ip")
		res, err := i.reuseStsIP(ctx, *addArgs.IfName, pod, ed)
		if err != nil {
			logging.Errorf("reuse statefulset ip failed:%v", err)
			return nil, err
		}

		if res != nil {
			return res, nil
		}
	} else {
		logging.Debugf("reuse the existing IP")
		res, err := i.reuseExistingIP(ctx, string(pod.UID), *addArgs.IfName, ed)
		if err != nil {
			logging.Errorf("reuse exist ip failed:%v", err)
			return nil, err
		}
		if res != nil {
			return res, nil
		}
	}
	//allocate ip in the normal mod
	res, err := i.allocateIps(ctx, addArgs, pod, ed, owner)
	if err != nil {
		logging.Errorf("allocate ip failed:%v", err)
		return nil, err
	}
	return res, err
}

// Delete release the ip with the given param
func (i *ipam) Delete(ctx context.Context, delArgs *models.IpamDelArgs) error {
	pod, err := i.podManager.GetPodByName(ctx, *delArgs.PodNamespace, *delArgs.PodName, true)
	if err != nil {
		logging.Errorf("get pod failed:%v", err)
		return err
	}

	//check is the pod alive
	alive := podmanager.IsPodAlive(pod)
	if !alive {
		logging.Errorf("pod %s/%s is still alive", pod.Namespace, pod.Name)
		return nil
	}

	//set timeout

	var timeout int64
	if pod != nil && pod.DeletionGracePeriodSeconds != nil {
		if *pod.DeletionGracePeriodSeconds != 0 {
			timeout = *pod.DeletionGracePeriodSeconds
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
			defer cancel()
		}
	}

	//get endpoint
	ed, err := i.endpointManager.GetEndpointByName(ctx, *delArgs.PodNamespace, *delArgs.PodName, false)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logging.Errorf("can not find the endpoint:%s/%s", *delArgs.PodNamespace, *delArgs.PodName)
			return nil
		}
		return logging.Errorf("find ipam endpoint failed :%v", err)
	}

	if ed != nil {
		logging.Debugf("get endpoint %s/%s", pod.Namespace, pod.Name)
	} else {
		logging.Errorf("find no endpoint")
	}

	//release ip
	err = i.releaseIP(ctx, *delArgs.PodUID, *delArgs.IfName, ed)
	if err != nil {
		logging.Errorf("release ip failed:%+v", err)
		return err
	}
	return nil
}

// Start will start the IPAM instance
func (i *ipam) Start(ctx context.Context) error {
	return nil
}
