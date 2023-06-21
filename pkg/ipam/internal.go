package ipam

import (
	"context"
	"fmt"
	"github.com/Inspur-Data/ipamwrapper/api/v1/models"
	"github.com/Inspur-Data/ipamwrapper/pkg/constant"
	inspuripamv1 "github.com/Inspur-Data/ipamwrapper/pkg/k8s/api/v1"
	"github.com/Inspur-Data/ipamwrapper/pkg/logging"
	"github.com/Inspur-Data/ipamwrapper/pkg/types"
	"github.com/Inspur-Data/ipamwrapper/pkg/utils/convert"
	corev1 "k8s.io/api/core/v1"
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
		return nil, fmt.Errorf("failed to patch IP allocation results to Endpoint: %v", err)
	}

	resIPs, resRoutes := convert.ConvertResultsToIPConfigsAndAllRoutes(res)
	addResp := &models.IpamAllocResponse{
		Ips:    resIPs,
		Routes: resRoutes,
	}
	logging.Debugf("allocate ip success: %v", addResp)

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
	// "ipam.spidernet.io/defaultv4ippool" and "ipam.spidernet.io/defaultv6ippool".
	ippools, err := i.getNsDefaultIPPool(ctx, pod.Namespace, *addArgs.IfName, true)
	if err == nil && ippools != nil {
		return ippools, nil
	} else {
		logging.Errorf("get ns default ippool failed:%v", err)
	}

	//todo add default ippools in the add args
	//get the default ippool from netconf
	ippools, err = i.getDefaultIPPoolFromNetconf(ctx, *addArgs.IfName, nil, nil, true)
	if err == nil && ippools != nil {
		return ippools, nil
	} else {
		logging.Errorf("get default ippool from netconf failed:%v", err)
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

	v4IppoolsMap := make(map[string]*inspuripamv1.IPPool)
	v6IppoolsMap := make(map[string]*inspuripamv1.IPPool)
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
			v4IppoolsMap[v4pool] = ippool
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

			if *ippool.Spec.IPVersion != constant.IPv4 {
				logging.Errorf("ippool:%s is not ipv4", v6pool)
				continue
			}
			v6IppoolsMap[v6pool] = ippool
		}
	}

	//todo Nodeaffinity namespace affinity

	//todo concurrent allocate !!!!!
	var result []*types.AllocationResult
	for name, v4ippool := range v4IppoolsMap {
		ip, err := i.ippoolManager.AllocateIP(ctx, v4ippool, nic, pod)
		if err != nil {
			logging.Errorf("allocate from ipool:%s failed:%v", name, err)
			continue
		}
		res := &types.AllocationResult{
			IP:           ip,
			Routes:       convert.ConvertSpecRoutesToOAIRoutes(nic, v4ippool.Spec.Routes),
			CleanGateway: cleanGateway,
		}
		result = append(result, res)
		break
	}

	return result, nil
}
