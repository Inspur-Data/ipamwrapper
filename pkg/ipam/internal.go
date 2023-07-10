package ipam

import (
	"context"
	"fmt"
	"github.com/Inspur-Data/ipamwrapper/api/v1/models"
	"github.com/Inspur-Data/ipamwrapper/pkg/constant"
	inspuripamv1 "github.com/Inspur-Data/ipamwrapper/pkg/k8s/api/v1"
	"github.com/Inspur-Data/ipamwrapper/pkg/logging"
	"github.com/Inspur-Data/ipamwrapper/pkg/manager/endpointmanager"
	"github.com/Inspur-Data/ipamwrapper/pkg/types"
	"github.com/Inspur-Data/ipamwrapper/pkg/utils/convert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sync"
)

// reuseStsIP will return IP from the endpoint
func (i *ipam) reuseStsIP(ctx context.Context, nic string, pod *corev1.Pod, endpoint *inspuripamv1.IPAMEndpoint) (*models.IpamAllocResponse, error) {
	valid := i.endpointManager.IsValidEndpoint(string(pod.UID), nic, endpoint, true)
	if !valid {
		logging.Debugf("endpoint is invalid, try to allocate IP in standard mode")
		return nil, nil
	}

	if err := i.updateIPPoolIPRecords(ctx, string(pod.UID), endpoint); err != nil {
		logging.Errorf("update ippool records failed:%v", err)
		return nil, err
	}

	if err := i.endpointManager.UpdateEndpoint(ctx, string(pod.UID), pod.Spec.NodeName, endpoint); err != nil {
		return nil, logging.Errorf("failed to update the current IP allocation of StatefulSet: %v", err)
	}

	ips, routes := convert.ConvertIPDetailsToIPsAndRoutes(endpoint.Status.IPs)
	addResp := &models.IpamAllocResponse{
		Ips:    ips,
		Routes: routes,
	}
	logging.Debugf("Succeed to reuse the IP of StatefulSet: %+v", *addResp)

	return addResp, nil
}

func (i *ipam) updateIPPoolIPRecords(ctx context.Context, uid string, endpoint *inspuripamv1.IPAMEndpoint) error {

	poolgroups := convert.GroupIPAllocationDetails(uid, endpoint.Status.IPs)
	//todo add metrics
	/*
		tickets := pius.Pools()
		timeRecorder := metric.NewTimeRecorder()
		if err := i.ipamLimiter.AcquireTicket(ctx, tickets...); err != nil {
			return fmt.Errorf("failed to queue correctly: %v", err)
		}
		defer i.ipamLimiter.ReleaseTicket(ctx, tickets...)

		// Record the metric of queuing time for allocating.
		metric.IPAMDurationConstruct.RecordIPAMAllocationLimitDuration(ctx, timeRecorder.SinceInSeconds())*/
	errCh := make(chan error, len(poolgroups))
	wg := sync.WaitGroup{}
	wg.Add(len(poolgroups))

	for pool, ipuid := range poolgroups {
		go func(poolName string, ipAndUIDs []types.IPAndUID) {
			defer wg.Done()

			if err := i.ippoolManager.UpdateAllocatedIPs(ctx, poolName, ipAndUIDs); err != nil {
				logging.Errorf("update allocated ips failed:%v", err)
				errCh <- err
				return
			}
			logging.Debugf("succeed to allocate IP addresses %+v from IPPool %s", ipAndUIDs, poolName)
		}(pool, ipuid)
	}
	wg.Wait()
	close(errCh)

	var errs []error
	for err := range errCh {
		errs = append(errs, err)
	}

	if len(errs) != 0 {
		return logging.Errorf("failed to re-allocate all allocated IP addresses %+v: %v", poolgroups, errs)
	}

	return nil
}

func (i *ipam) reuseExistingIP(ctx context.Context, uid, nic string, endpoint *inspuripamv1.IPAMEndpoint) (*models.IpamAllocResponse, error) {
	//todo check create-delete-create scene
	// Create -> Delete -> Create a Pod with the same namespace and name in
	// a short time will cause some unexpected phenomena discussed in
	// https://github.com/spidernet-io/spiderpool/issues/1187.
	if endpoint != nil && endpoint.Status.UID != uid {
		return nil, fmt.Errorf("currently, the IP allocation of the Pod %s/%s (UID: %s) is being recycled. You may create two Pods with the same namespace and name in a very short time", endpoint.Namespace, endpoint.Name, endpoint.Status.UID)
	}

	valid := i.endpointManager.IsValidEndpoint(uid, nic, endpoint, false)
	if !valid {
		logging.Debugf("endpoint is invalid")
		return nil, nil
	}

	ips, routes := convert.ConvertIPDetailsToIPsAndRoutes(endpoint.Status.IPs)
	addResp := &models.IpamAllocResponse{
		Ips:    ips,
		Routes: routes,
	}
	logging.Debugf("Succeed to retrieve the IP allocation: %+v", *addResp)

	return addResp, nil
}

// allocateIps is the standard mod about ip allocation
func (i *ipam) allocateIps(ctx context.Context, addArgs *models.IpamAllocArgs, pod *corev1.Pod, endpoint *inspuripamv1.IPAMEndpoint, podCtl types.PodTopController) (*models.IpamAllocResponse, error) {

	/*
		customRoutes, err := getRouteFromAnno(pod)
		if err != nil {
			return nil, err
		}*/

	ippoolCandidate, err := i.getCandidatePool(ctx, addArgs, pod, podCtl)
	if err != nil {
		return nil, logging.Errorf("get candidate ippool failed:%v", err)
	}
	if ippoolCandidate == nil {
		return nil, logging.Errorf("ipppool candidate is nill")
	}

	res, err := i.allocateIPsFromAllCandidates(ctx, *addArgs.IfName, ippoolCandidate, pod, true)
	if err != nil || res == nil {
		return nil, logging.Errorf("allocate ip from candidate failed:%v", err)
	}

	//todo group route
	/*
		if err = groupCustomRoutes(ctx, customRoutes, results); err != nil {
			return nil, fmt.Errorf("failed to group custom routes %+v: %v", customRoutes, err)
		}*/

	if err = i.endpointManager.PatchIPAllocationResults(ctx, res, endpoint, pod, podCtl); err != nil {
		//todo ippool status rollback
		return nil, fmt.Errorf("failed to patch IP allocation results to Endpoint: %v", err)
	}

	resIPs, resRoutes := convert.ConvertResultsToIPConfigsAndAllRoutes(res)
	addResp := &models.IpamAllocResponse{
		Ips:    resIPs,
		Routes: resRoutes,
	}
	logging.Debugf("allocate ip success: %v", *addResp)

	return addResp, nil
}

// getCandidatePool will return the ippool candidate after a series of filter
func (i *ipam) getCandidatePool(ctx context.Context, addArgs *models.IpamAllocArgs, pod *corev1.Pod, podController types.PodTopController) (*types.AnnoPodIPPoolValue, error) {
	//todo subnet has the highest order

	//get ippool from annotation "ipam.inspur.io/ippool"
	if anno, ok := pod.Annotations[constant.AnnoPodIPPool]; ok {
		return getCandidatePoolFromAnno(anno, *addArgs.IfName, true)
	}

	// get ippool from namespace annotations
	// "ipam.inspur.io/defaultv4ippool" and "ipam.inspur.io/defaultv6ippool".
	ippools, err := i.getNsDefaultIPPool(ctx, pod.Namespace, *addArgs.IfName, true)
	if err == nil && ippools != nil {
		return ippools, nil
	}

	//todo add default ippools in the add args

	//get the default ippool from netconf
	ippools, err = i.getDefaultIPPoolFromNetconf(ctx, *addArgs.IfName, nil, nil, true)
	if err == nil && ippools != nil {
		return ippools, nil
	}

	//get the default ippool
	ippools, err = i.getDefaultIPPool(ctx, *addArgs.IfName, true)
	if err == nil && ippools != nil {
		return ippools, nil
	} else {
		logging.Errorf("get default ippool failed:%v", err)
	}

	return nil, nil
}

// getNsDefaultIPPool get default ippool from namespace
func (i *ipam) getNsDefaultIPPool(ctx context.Context, namespace, nic string, cleanGateway bool) (*types.AnnoPodIPPoolValue, error) {
	ns, err := i.nsManager.GetNamespace(ctx, namespace, constant.UseCache)
	if err != nil {
		logging.Errorf("get namespace failed:%v", err)
		return nil, err
	}
	nsDefaultV4Pools, nsDefaultV6Pools, err := i.nsManager.GetNSDefaultPools(ns)
	if err != nil {
		logging.Errorf("get namespace default ippool failed:%v", err)
		return nil, err
	}

	if len(nsDefaultV4Pools) == 0 && len(nsDefaultV6Pools) == 0 {
		logging.Errorf("ipv4 and ipv6 ippool is nil")
		return nil, nil
	}

	ippools := types.AnnoPodIPPoolValue{}
	ippools.IPv4Pools = nsDefaultV4Pools
	ippools.IPv6Pools = nsDefaultV6Pools
	return &ippools, nil
}

// getDefaultIPPoolFromNetconf get the ippool from args
func (i *ipam) getDefaultIPPoolFromNetconf(ctx context.Context, nic string, defaultIPv4Pool, defaultIPv6Pool []string, cleanGateway bool) (*types.AnnoPodIPPoolValue, error) {
	if len(defaultIPv4Pool) == 0 && len(defaultIPv6Pool) == 0 {
		logging.Errorf("ipv4 and ipv6 ippool is nil")
		return nil, nil
	}
	ippools := types.AnnoPodIPPoolValue{}
	ippools.IPv4Pools = defaultIPv4Pool
	ippools.IPv6Pools = defaultIPv6Pool
	return &ippools, nil
}

// getDefaultIPPool get the ippool from cluster who has the default spec
func (i *ipam) getDefaultIPPool(ctx context.Context, nic string, cleanGateway bool) (*types.AnnoPodIPPoolValue, error) {

	ipPoolList, err := i.ippoolManager.ListIPPools(
		ctx,
		constant.UseCache,
		client.MatchingLabels{"default": "true"},
		//client.MatchingFields{"spec.default": strconv.FormatBool(true)},
	)
	if err != nil {
		logging.Errorf("list ippool failed:%v", err)
		return nil, err
	}

	if len(ipPoolList.Items) == 0 {
		return nil, logging.Errorf("no ippool has the default spec")
	}

	ippools := types.AnnoPodIPPoolValue{}
	var v4Pools []string
	var v6Pools []string
	for _, ipPool := range ipPoolList.Items {
		if *ipPool.Spec.IPVersion == constant.IPv4 {
			v4Pools = append(v4Pools, ipPool.Name)
		} else {
			v6Pools = append(v6Pools, ipPool.Name)
		}
	}

	ippools.IPv4Pools = v4Pools
	ippools.IPv6Pools = v6Pools

	return &ippools, nil
}

// allocateIPsFromAllCandidates allocate ips from the candidate
func (i *ipam) allocateIPsFromAllCandidates(ctx context.Context, nic string, ippools *types.AnnoPodIPPoolValue, pod *corev1.Pod, cleanGateway bool) ([]*types.AllocationResult, error) {
	//checkout ip version
	if i.config.EnableIPv4 && len(ippools.IPv4Pools) == 0 {
		return nil, logging.Errorf("ipv4 enabled but ipv4 ipools is nil")
	}

	if i.config.EnableIPv6 && len(ippools.IPv6Pools) == 0 {
		return nil, logging.Errorf("ipv6 enabled but ipv6 ipools is nil")
	}

	IppoolsMap := make(map[string]*inspuripamv1.IPPool)

	for _, v4pool := range ippools.IPv4Pools {
		ippool, err := i.ippoolManager.GetIPPoolByName(ctx, v4pool, false)
		if err != nil {
			logging.Errorf("get ippool:%s failed :%v", v4pool, err)
			continue
		} else {
			if ippool.DeletionTimestamp != nil {
				logging.Errorf("ippool:%s is deleting", v4pool)
				continue
			}

			if *ippool.Spec.IPVersion != constant.IPv4 {
				logging.Errorf("ippool:%s is not ipv4", v4pool)
				continue
			}
			IppoolsMap[v4pool] = ippool
		}
	}

	for _, v6pool := range ippools.IPv6Pools {
		ippool, err := i.ippoolManager.GetIPPoolByName(ctx, v6pool, true)
		if err != nil {
			logging.Errorf("get ippool:%s failed :%v", v6pool, err)
			continue
		} else {
			if ippool.DeletionTimestamp != nil {
				logging.Errorf("ippool:%s is deleting", v6pool)
				continue
			}

			if *ippool.Spec.IPVersion != constant.IPv6 {
				logging.Errorf("ippool:%s is not ipv6", v6pool)
				continue
			}
			IppoolsMap[v6pool] = ippool
		}
	}

	//todo Nodeaffinity namespace affinity
	for poolname, pool := range IppoolsMap {
		//node affinity
		if pool.Spec.NodeAffinity != nil {
			node, err := i.nodeManager.GetNodeByName(ctx, pod.Spec.NodeName, constant.UseCache)
			if err != nil {
				delete(IppoolsMap, poolname)
				continue
			}
			selector, err := metav1.LabelSelectorAsSelector(pool.Spec.NodeAffinity)
			if err != nil {
				delete(IppoolsMap, poolname)
				continue
			}
			if !selector.Matches(labels.Set(node.Labels)) {
				logging.Errorf("unmatched node affinity of IPPool %s", poolname)
				delete(IppoolsMap, poolname)
			}
		}

		//namespace affinity
		if pool.Spec.NamespaceAffinity != nil {
			namespace, err := i.nsManager.GetNamespace(ctx, pod.Namespace, constant.UseCache)
			if err != nil {
				delete(IppoolsMap, poolname)
				continue
			}
			selector, err := metav1.LabelSelectorAsSelector(pool.Spec.NamespaceAffinity)
			if err != nil {
				delete(IppoolsMap, poolname)
				continue
			}
			if !selector.Matches(labels.Set(namespace.Labels)) {
				logging.Errorf("unmatched namespace affinity of IPPool %s", poolname)
				delete(IppoolsMap, poolname)
			}
		}
	}

	if len(IppoolsMap) == 0 {
		return nil, logging.Errorf("all ippool candidate are invalid")
	}

	//todo concurrent allocate !!!!!

	var result []*types.AllocationResult
	var errs []error
	for name, ippool := range IppoolsMap {
		ip, err := i.ippoolManager.AllocateIP(ctx, ippool, nic, pod, i.config.IPv4ReservedIP, i.config.IPv6ReservedIP)
		if err != nil {
			logging.Errorf("allocate from ipool:%s failed:%v", name, err)
			errs = append(errs, err)
			continue
		}
		res := &types.AllocationResult{
			IP:           ip,
			Routes:       convert.ConvertSpecRoutesToOAIRoutes(nic, ippool.Spec.Routes),
			CleanGateway: cleanGateway,
		}
		result = append(result, res)
		break
	}

	if len(errs) == len(IppoolsMap) {
		return nil, logging.Errorf("allocate ip from all ippools failed")
	}

	return result, nil
}

func (i *ipam) releaseIP(ctx context.Context, uid, nic string, endpoint *inspuripamv1.IPAMEndpoint) error {

	// judge whether an sts needs to release its currently allocated IP addresses.
	if i.config.EnableStatefulSet && endpoint.Status.TopOwner == constant.KindStatefulSet {
		valid, err := i.stsManager.IsValidStsPod(ctx, endpoint.Namespace, endpoint.Name, endpoint.Status.TopOwner)
		if nil != err {
			return fmt.Errorf("failed to check pod %s/%s whether is a valid StatefulSet pod: %v", endpoint.Namespace, endpoint.Name, err)
		}

		if valid {
			logging.Errorf("no need to release the pod ip")
			return nil
		}

	}

	//delete ipam endpoint
	if err := i.endpointManager.DeleteEndpoint(ctx, endpoint); err != nil {
		logging.Errorf("delete endpoint failed:%v", err)
		return err
	}

	allocation := endpointmanager.GetEndpointIP(uid, nic, endpoint, false)
	if allocation == nil {
		logging.Debugf("this endpoint cant hanve ip allocation ")
		return nil
	}

	logging.Debugf("release IP allocation details: %+v", allocation.IPs)
	if err := i.release(ctx, allocation.UID, allocation.IPs); err != nil {
		return err
	}

	if err := i.endpointManager.RemoveFinalizer(ctx, endpoint); err != nil {
		return fmt.Errorf("failed to clean Endpoint: %v", err)
	}

	return nil
}

func (i *ipam) release(ctx context.Context, uid string, details []inspuripamv1.IPAllocationDetail) error {
	//group the ip allocation detail to map. key is ippool name,value is IP and UID
	detailGroups := convert.GroupIPAllocationDetails(uid, details)
	//todo add metrics
	/*
			tickets := pius.Pools()
			timeRecorder := metric.NewTimeRecorder()
			if err := i.ipamLimiter.AcquireTicket(ctx, tickets...); err != nil {
				return fmt.Errorf("failed to queue correctly: %v", err)
			}
			defer i.ipamLimiter.ReleaseTicket(ctx, tickets...)

		// Record the metric of queuing time for release.
		metric.IPAMDurationConstruct.RecordIPAMReleaseLimitDuration(ctx, timeRecorder.SinceInSeconds())
	*/
	errCh := make(chan error, len(detailGroups))
	wg := sync.WaitGroup{}
	wg.Add(len(detailGroups))

	for p, detail := range detailGroups {
		go func(poolName string, ipAndUIDs []types.IPAndUID) {
			defer wg.Done()

			if err := i.ippoolManager.ReleaseIP(ctx, poolName, ipAndUIDs); err != nil {
				logging.Errorf("release ip failed:%v", err)
				errCh <- err
				return
			}
			logging.Debugf("release IP successful ippool name:%s,ip:+%v", poolName, ipAndUIDs)
		}(p, detail)
	}
	wg.Wait()
	close(errCh)

	var errs []error
	for err := range errCh {
		errs = append(errs, err)
	}

	if len(errs) != 0 {
		return logging.Errorf("failed to release all allocated IP addresses %+v", detailGroups)
	}

	return nil
}
