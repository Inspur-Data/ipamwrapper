package ipam

import (
	"context"
	"github.com/Inspur-Data/ipamwrapper/api/v1/models"
	inspuripamv1 "github.com/Inspur-Data/ipamwrapper/pkg/k8s/api/v1"
	"github.com/Inspur-Data/ipamwrapper/pkg/logging"
	"github.com/Inspur-Data/ipamwrapper/pkg/manager/endpointmanager"
	"github.com/Inspur-Data/ipamwrapper/pkg/types"
	"github.com/Inspur-Data/ipamwrapper/pkg/utils/convert"
	corev1 "k8s.io/api/core/v1"
	"sync"
)

// reuseStsIP will return IP from the endpoint
func (i *ipam) reuseStsIP(ctx context.Context, nic string, pod *corev1.Pod, endpoint *inspuripamv1.IPAMEndpoint) (*models.IpamAllocResponse, error) {
	allocation := endpointmanager.GetEndpointIP(string(pod.UID), nic, endpoint, true)
	if allocation == nil {
		// this is the first allocation or multi-NIC.
		logging.Debugf("ip allocation is not found, try to allocate IP in standard mode")
		return nil, nil
	}

	if err := i.updateIPPoolIPRecords(ctx, string(pod.UID), endpoint); err != nil {
		logging.Errorf("reallocate ip failed:%v", err)
		return nil, err
	}

	if err := i.endpointManager.ReallocateCurrentIPAllocation(ctx, string(pod.UID), pod.Spec.NodeName, endpoint); err != nil {
		return nil, logging.Errorf("failed to update the current IP allocation of StatefulSet: %v", err)
	}

	ips, routes := convert.ConvertIPDetailsToIPConfigsAndAllRoutes(endpoint.Status.IPs)
	addResp := &models.IpamAllocResponse{
		Ips:    ips,
		Routes: routes,
	}
	logging.Debugf("Succeed to reuse the IP of StatefulSet: %+v", *addResp)

	return addResp, nil
}

func (i *ipam) updateIPPoolIPRecords(ctx context.Context, uid string, endpoint *inspuripamv1.IPAMEndpoint) error {

	pius := convert.GroupIPAllocationDetails(uid, endpoint.Status.IPs)
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
	errCh := make(chan error, len(pius))
	wg := sync.WaitGroup{}
	wg.Add(len(pius))

	for p, ius := range pius {
		go func(poolName string, ipAndUIDs []types.IPAndUID) {
			defer wg.Done()

			if err := i.ippoolManager.UpdateAllocatedIPs(ctx, poolName, ipAndUIDs); err != nil {
				logging.Errorf("update allocated ips failed:%v", err)
				errCh <- err
				return
			}
			logging.Debugf("succeed to allocate IP addresses %+v from IPPool %s", ipAndUIDs, poolName)
		}(p, ius)
	}
	wg.Wait()
	close(errCh)

	var errs []error
	for err := range errCh {
		errs = append(errs, err)
	}

	if len(errs) != 0 {
		return logging.Errorf("failed to re-allocate all allocated IP addresses %+v: %v", pius, errs)
	}

	return nil
}
