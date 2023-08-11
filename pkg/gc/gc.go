// Copyright 2022 Authors of Inspur
// SPDX-License-Identifier: Apache-2.0
package gc

import (
	"context"
	"github.com/Inspur-Data/ipamwrapper/pkg/constant"
	inspuripamv1 "github.com/Inspur-Data/ipamwrapper/pkg/k8s/api/v1"
	"github.com/Inspur-Data/ipamwrapper/pkg/logging"
	"github.com/Inspur-Data/ipamwrapper/pkg/manager/endpointmanager"
	"github.com/Inspur-Data/ipamwrapper/pkg/manager/ippoolmanager"
	"github.com/Inspur-Data/ipamwrapper/pkg/manager/nodemanager"
	"github.com/Inspur-Data/ipamwrapper/pkg/manager/nsmanager"
	"github.com/Inspur-Data/ipamwrapper/pkg/manager/podmanager"
	"github.com/Inspur-Data/ipamwrapper/pkg/manager/stsmanager"
	"github.com/Inspur-Data/ipamwrapper/pkg/types"
	"github.com/Inspur-Data/ipamwrapper/pkg/utils/convert"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	ctrl "sigs.k8s.io/controller-runtime"
	"strings"
	"sync"
	"time"
)

type GC interface {
	StartGC(ctx context.Context, period int)
}

type gc struct {
	podManager      podmanager.PodManager
	endpointManager endpointmanager.EndpointManager
	ippoolManager   ippoolmanager.IPPoolManager
	nsManager       nsmanager.NsManager
	stsManager      stsmanager.StatefulSetManager
	nodeManager     nodemanager.NodeManager
}

func NewGC(manager ctrl.Manager) (GC, error) {

	nsMgr, err := nsmanager.NewNamespaceManager(
		manager.GetClient(),
		manager.GetAPIReader(),
	)
	if err != nil {
		logging.Panicf("init namespace manager failed:%v", err)
	}

	podMgr, err := podmanager.NewPodManager(
		manager.GetClient(),
		manager.GetAPIReader(),
	)
	if err != nil {
		logging.Panicf("init pod manager failed:%v", err)
	}

	endpointMgr, err := endpointmanager.NewEndpointManager(
		manager.GetClient(),
		manager.GetAPIReader(),
	)
	if err != nil {
		logging.Panicf("init endpoint manager failed:%v", err)
	}

	ipPoolMgr, err := ippoolmanager.NewIPPoolManager(
		ippoolmanager.MgrConfig{},
		manager.GetClient(),
		manager.GetAPIReader(),
	)
	if err != nil {
		logging.Panicf("init ippool manager failed: %v", err)
	}

	stsMgr, err := stsmanager.NewStatefulSetManager(
		manager.GetClient(),
		manager.GetAPIReader(),
	)
	if err != nil {
		logging.Panicf("init ippool manager failed: %v", err)
	}

	nodeMgr, err := nodemanager.NewNodeManager(
		manager.GetClient(),
		manager.GetAPIReader(),
	)
	if err != nil {
		logging.Panicf("init ippool manager failed: %v", err)
	}

	return &gc{
		podManager:      podMgr,
		endpointManager: endpointMgr,
		ippoolManager:   ipPoolMgr,
		nsManager:       nsMgr,
		stsManager:      stsMgr,
		nodeManager:     nodeMgr,
	}, nil
}

func (g *gc) StartGC(ctx context.Context, period int) {
	logging.Debugf("Start the gc period...")
	duration := time.Duration(period) * time.Second
	timer := time.NewTimer(duration)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			g.doGC(ctx)
		case <-ctx.Done():
			logging.Debugf("receive ctx done, stop monitoring gc signal!")
			return
		}
		timer.Reset(time.Duration(period) * time.Second)
	}
}

func (g *gc) doGC(ctx context.Context) {
	logging.Debugf("start do gc....")
	poolList, err := g.ippoolManager.ListIPPools(ctx, constant.UseCache)
	if nil != err {
		if apierrors.IsNotFound(err) {
			logging.Warningf("list ippool failed, not found!")
			return
		}
		logging.Errorf("get ipppool list failed: '%v'", err)
		return
	}

	var v4poolList, v6poolList []inspuripamv1.IPPool
	for i := range poolList.Items {
		if poolList.Items[i].Spec.IPVersion != nil {
			if *poolList.Items[i].Spec.IPVersion == constant.IPv4 {
				v4poolList = append(v4poolList, poolList.Items[i])
			} else {
				v6poolList = append(v6poolList, poolList.Items[i])
			}
		}
	}

	wg := sync.WaitGroup{}
	if len(v4poolList) != 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			g.checkIppool(ctx, v4poolList)
		}()
	}

	if len(v6poolList) != 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			g.checkIppool(ctx, v6poolList)
		}()
	}

	wg.Wait()
}

func (g *gc) checkIppool(ctx context.Context, pools []inspuripamv1.IPPool) {
	for _, pool := range pools {
		allocatedIPs, err := convert.UnmarshalIPPoolAllocatedIPs(pool.Status.AllocatedIPs)
		if nil != err {
			logging.Errorf("failed to parse IPPool '%v' allocatedIPs, error: %v", pool, err)
			continue
		}

		for ip, ipDetail := range allocatedIPs {
			podNS, podName, err := cache.SplitMetaNamespaceKey(ipDetail.NamespacedName)
			if err != nil {
				logging.Errorf(err.Error())
				continue
			}
			pod, err := g.podManager.GetPodByName(ctx, podNS, podName, constant.UseCache)
			if err != nil {
				if apierrors.IsNotFound(err) {
					endpoint, err := g.endpointManager.GetEndpointByName(ctx, podNS, podName, constant.UseCache)
					if nil != err {
						// just continue if we meet other errors
						if !apierrors.IsNotFound(err) {
							logging.Errorf("failed to get IPAM Endpoint: %v", err)
							continue
						}
					} else {
						if endpoint.Status.TopOwner == constant.KindStatefulSet {
							isValidStsPod, err := g.stsManager.IsValidStsPod(ctx, podNS, podName, constant.KindStatefulSet)
							if nil != err {
								logging.Errorf("failed to check StatefulSet pod IP '%s' should be cleaned or not, error: %v", ip, err)
								continue
							}
							if isValidStsPod {
								logging.Debugf("no deed to release IP '%s' for StatefulSet pod ", ip)
								continue
							}
						}
					}

					err = g.releaseIP(ctx, pool.Name, ip, ipDetail)
					if nil != err {
						logging.Errorf(err.Error())
					}
					// continue to the next poolIP
					continue
				}
				logging.Errorf("failed to get pod from kubernetes, error '%v'", err)
				continue
			} else {
				//check Pod information
				g.checkPod(ctx, pod, pool.Name, ip, pool.Spec.IPVersion, ipDetail)
			}
		}
	}
}

func (g *gc) releaseIP(ctx context.Context, poolName, poolIP string, poolIPAllocation inspuripamv1.PoolIPAllocation) error {
	singleIP := []types.IPAndUID{{IP: poolIP, UID: poolIPAllocation.PodUID}}
	err := g.ippoolManager.ReleaseIP(ctx, poolName, singleIP)
	if nil != err {
		return logging.Errorf("failed to release IP '%s', error: '%v'", poolIP, err)
	}
	podNS, podName, err := cache.SplitMetaNamespaceKey(poolIPAllocation.NamespacedName)
	if err != nil {
		return err
	}

	endpoint, err := g.endpointManager.GetEndpointByName(ctx, podNS, podName, constant.UseCache)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logging.Debugf("Endpoint '%s/%s' has already delete", podNS, podName)
			return nil
		}
		return err
	}

	if err := g.endpointManager.RemoveFinalizer(ctx, endpoint); err != nil {
		return err
	}
	return nil
}

func (g *gc) checkPod(ctx context.Context, pod *v1.Pod, poolName, poolIP string, ipversion *int64, poolIPAllocation inspuripamv1.PoolIPAllocation) error {
	//checkout hostNetwork
	if pod.Spec.HostNetwork {
		logging.Debugf("pod with hostnetwork should not be process")
		return nil
	}

	//check stateful set's pod
	ownerRef := metav1.GetControllerOf(pod)
	if ownerRef != nil && ownerRef.Kind == constant.KindStatefulSet {
		isValidStsPod, err := g.stsManager.IsValidStsPod(context.TODO(), pod.Namespace, pod.Name, ownerRef.Kind)
		if nil != err {
			return err
		}

		// StatefulSet pod restarted, no need to process it.
		if isValidStsPod {
			return logging.Errorf("no need to process the valid sts pod")
		}
	}

	//check the uid
	if string(pod.UID) != poolIPAllocation.PodUID {
		singleIP := []types.IPAndUID{{IP: poolIP, UID: poolIPAllocation.PodUID}}
		err := g.ippoolManager.ReleaseIP(ctx, poolName, singleIP)
		if nil != err {
			return logging.Errorf("failed to release IP '%s', error: '%v'", poolIP, err)
		}
		logging.Debugf("pod uid has changed,will delete the ipam endpoint")
		endpoint, err := g.endpointManager.GetEndpointByName(ctx, pod.Namespace, pod.Name, constant.UseCache)
		if err != nil {
			if apierrors.IsNotFound(err) {
				logging.Debugf("Endpoint '%s/%s' has already delete", pod.Namespace, pod.Name)
				return nil
			}
			return err
		}

		if err := g.endpointManager.RemoveFinalizer(ctx, endpoint); err != nil {
			return err
		}
	} else {
		endpoint, err := g.endpointManager.GetEndpointByName(ctx, pod.Namespace, pod.Name, constant.UseCache)
		if err != nil {
			if apierrors.IsNotFound(err) {
				logging.Debugf("Endpoint '%s/%s' has already delete", pod.Namespace, pod.Name)
				return nil
			}
			return err
		}

		needRelease := true
		for _, ipdetail := range endpoint.Status.IPs {
			if *ipversion == constant.IPv4 {
				if ipdetail.IPv4 != nil && strings.Split(*ipdetail.IPv4, "/")[0] == poolIP {
					needRelease = false
				}
			} else {
				if ipdetail.IPv6 != nil && strings.Split(*ipdetail.IPv6, "/")[0] == poolIP {
					needRelease = false
				}
			}
		}

		if needRelease {
			singleIP := []types.IPAndUID{{IP: poolIP, UID: poolIPAllocation.PodUID}}
			err := g.ippoolManager.ReleaseIP(ctx, poolName, singleIP)
			if nil != err {
				return logging.Errorf("failed to release IP '%s', error: '%v'", poolIP, err)
			}
		}
	}
	return nil
}
