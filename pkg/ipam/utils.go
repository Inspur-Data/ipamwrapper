package ipam

import (
	"encoding/json"
	"github.com/Inspur-Data/ipamwrapper/api/v1/models"
	"github.com/Inspur-Data/ipamwrapper/pkg/logging"
	"github.com/Inspur-Data/ipamwrapper/pkg/types"
	"github.com/Inspur-Data/ipamwrapper/pkg/utils/convert"
	corev1 "k8s.io/api/core/v1"
)

// getRouteFromAnno get route from pod's annotation
func getRouteFromAnno(pod *corev1.Pod) ([]*models.Route, error) {
	//todo add routes to the constant
	anno, ok := pod.Annotations["routes"]
	if !ok {
		return nil, nil
	}

	var annoRoutes types.AnnoPodRoutes
	err := json.Unmarshal([]byte(anno), &annoRoutes)
	if err != nil {
		logging.Errorf("json unmarshal anno:%v failed: %v", anno, err)
		return nil, err
	}

	//todo check the route is valid or not
	/*
		for _, route := range annoRoutes {
			if err := spiderpoolip.IsRouteWithoutIPVersion(route.Dst, route.Gw); err != nil {
				return nil, fmt.Errorf("%w: %v", errPrefix, err)
			}
		}*/

	return convert.ConvertAnnoRoutes(annoRoutes), nil

}

func getCandidatePoolFromAnno(anno, nic string, cleanGateway bool) (*types.AnnoPodIPPoolValue, error) {
	var annoPodIPPool types.AnnoPodIPPoolValue

	if err := json.Unmarshal([]byte(anno), &annoPodIPPool); err != nil {
		logging.Errorf("json unmarshal failed:%v", err)
		return nil, err
	}
	return &annoPodIPPool, nil
}

func getCandidatePoolFromNsAnno(anno, nic string, cleanGateway bool) (*types.AnnoPodIPPoolValue, error) {
	var annoPodIPPool types.AnnoPodIPPoolValue

	if err := json.Unmarshal([]byte(anno), &annoPodIPPool); err != nil {
		logging.Errorf("json unmarshal failed:%v", err)
		return nil, err
	}
	return &annoPodIPPool, nil
}
