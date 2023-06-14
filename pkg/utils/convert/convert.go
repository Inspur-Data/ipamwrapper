// Copyright 2022 Authors of Inspur
// SPDX-License-Identifier: Apache-2.0

package convert

import (
	"encoding/json"
	"net"
	"strings"

	"github.com/Inspur-Data/ipamwrapper/api/v1/models"
	"github.com/Inspur-Data/ipamwrapper/pkg/constant"
	inspuripamv1 "github.com/Inspur-Data/ipamwrapper/pkg/k8s/api/v1"
	"github.com/Inspur-Data/ipamwrapper/pkg/types"
	"github.com/asaskevich/govalidator"
)

func ConvertIPDetailsToIPConfigsAndAllRoutes(details []inspuripamv1.IPAllocationDetail) ([]*models.IPConfig, []*models.Route) {
	var ips []*models.IPConfig
	var routes []*models.Route
	for _, d := range details {
		nic := d.NIC
		if d.IPv4 != nil {
			version := constant.IPv4
			var ipv4Gateway string
			if d.IPv4Gateway != nil {
				ipv4Gateway = *d.IPv4Gateway
				routes = append(routes, genDefaultRoute(nic, ipv4Gateway))
			}
			ips = append(ips, &models.IPConfig{
				Address: d.IPv4,
				Gateway: ipv4Gateway,
				IPPool:  *d.IPv4Pool,
				Nic:     &nic,
				Version: &version,
			})
		}

		if d.IPv6 != nil {
			version := constant.IPv6
			var ipv6Gateway string
			if d.IPv6Gateway != nil {
				ipv6Gateway = *d.IPv6Gateway
				routes = append(routes, genDefaultRoute(nic, ipv6Gateway))
			}
			ips = append(ips, &models.IPConfig{
				Address: d.IPv6,
				Gateway: ipv6Gateway,
				IPPool:  *d.IPv6Pool,
				Nic:     &nic,
				Version: &version,
			})
		}

		routes = append(routes, ConvertSpecRoutesToOAIRoutes(d.NIC, d.Routes)...)
	}

	return ips, routes
}

func ConvertResultsToIPConfigsAndAllRoutes(results []*types.AllocationResult) ([]*models.IPConfig, []*models.Route) {
	var ips []*models.IPConfig
	var routes []*models.Route
	for _, r := range results {
		ips = append(ips, r.IP)
		routes = append(routes, r.Routes...)

		if r.CleanGateway {
			continue
		}

		if r.IP.Gateway != "" {
			routes = append(routes, genDefaultRoute(*r.IP.Nic, r.IP.Gateway))
		}
	}

	return ips, routes
}

func genDefaultRoute(nic, gateway string) *models.Route {
	var route *models.Route
	if govalidator.IsIPv4(gateway) {
		dst := "0.0.0.0/0"
		route = &models.Route{
			IfName: &nic,
			Dst:    &dst,
			Gw:     &gateway,
		}
	}

	if govalidator.IsIPv6(gateway) {
		dst := "::/0"
		route = &models.Route{
			IfName: &nic,
			Dst:    &dst,
			Gw:     &gateway,
		}
	}

	return route
}

func ConvertResultsToIPDetails(results []*types.AllocationResult) []inspuripamv1.IPAllocationDetail {
	nicToDetail := map[string]*inspuripamv1.IPAllocationDetail{}
	for _, r := range results {
		var gateway *string
		var cleanGateway *bool
		if r.IP.Gateway != "" {
			gateway = new(string)
			cleanGateway = new(bool)
			*gateway = r.IP.Gateway
			*cleanGateway = r.CleanGateway
		}

		address := *r.IP.Address
		pool := r.IP.IPPool
		routes := ConvertOAIRoutesToSpecRoutes(r.Routes)

		if d, ok := nicToDetail[*r.IP.Nic]; ok {
			if *r.IP.Version == constant.IPv4 {
				d.IPv4 = &address
				d.IPv4Pool = &pool
				d.IPv4Gateway = gateway
				d.Routes = append(d.Routes, routes...)
			} else {
				d.IPv6 = r.IP.Address
				d.IPv6Pool = &r.IP.IPPool
				d.IPv6Gateway = gateway
				d.Routes = append(d.Routes, routes...)
			}
			continue
		}

		if *r.IP.Version == constant.IPv4 {
			nicToDetail[*r.IP.Nic] = &inspuripamv1.IPAllocationDetail{
				NIC:          *r.IP.Nic,
				IPv4:         &address,
				IPv4Pool:     &pool,
				IPv4Gateway:  gateway,
				CleanGateway: cleanGateway,
				Routes:       routes,
			}
		} else {
			nicToDetail[*r.IP.Nic] = &inspuripamv1.IPAllocationDetail{
				NIC:          *r.IP.Nic,
				IPv6:         &address,
				IPv6Pool:     &pool,
				IPv6Gateway:  gateway,
				CleanGateway: cleanGateway,
				Routes:       routes,
			}
		}
	}

	details := []inspuripamv1.IPAllocationDetail{}
	for _, d := range nicToDetail {
		details = append(details, *d)
	}

	return details
}

func ConvertAnnoPodRoutesToOAIRoutes(annoPodRoutes types.AnnoPodRoutesValue) []*models.Route {
	var routes []*models.Route
	for _, r := range annoPodRoutes {
		dst := r.Dst
		gw := r.Gw
		routes = append(routes, &models.Route{
			IfName: new(string),
			Dst:    &dst,
			Gw:     &gw,
		})
	}

	return routes
}

func ConvertSpecRoutesToOAIRoutes(nic string, specRoutes []inspuripamv1.Route) []*models.Route {
	var routes []*models.Route
	for _, r := range specRoutes {
		dst := r.Dst
		gw := r.Gw
		routes = append(routes, &models.Route{
			IfName: &nic,
			Dst:    &dst,
			Gw:     &gw,
		})
	}

	return routes
}

func ConvertOAIRoutesToSpecRoutes(oaiRoutes []*models.Route) []inspuripamv1.Route {
	var routes []inspuripamv1.Route
	for _, r := range oaiRoutes {
		routes = append(routes, inspuripamv1.Route{
			Dst: *r.Dst,
			Gw:  *r.Gw,
		})
	}

	return routes
}

func GroupIPAllocationDetails(uid string, details []inspuripamv1.IPAllocationDetail) types.PoolNameToIPAndUIDs {
	pius := types.PoolNameToIPAndUIDs{}
	for _, d := range details {
		if d.IPv4 != nil {
			pius[*d.IPv4Pool] = append(pius[*d.IPv4Pool], types.IPAndUID{
				IP:  strings.Split(*d.IPv4, "/")[0],
				UID: uid,
			})
		}
		if d.IPv6 != nil {
			pius[*d.IPv6Pool] = append(pius[*d.IPv6Pool], types.IPAndUID{
				IP:  strings.Split(*d.IPv6, "/")[0],
				UID: uid,
			})
		}
	}

	return pius
}

func GenIPConfigResult(allocateIP net.IP, nic string, ipPool *inspuripamv1.IPPool) *models.IPConfig {
	/*
		ipNet, _ := spiderpoolip.ParseIP(*ipPool.Spec.IPVersion, ipPool.Spec.Subnet, true)
		ipNet.IP = allocateIP
		address := ipNet.String()

		var gateway string
		if ipPool.Spec.Gateway != nil {
			gateway = *ipPool.Spec.Gateway
		}

		return &models.IPConfig{
			Address: &address,
			Gateway: gateway,
			IPPool:  ipPool.Name,
			Nic:     &nic,
			Version: ipPool.Spec.IPVersion,
			Vlan:    *ipPool.Spec.Vlan,
		}*/
	return &models.IPConfig{}
}

func UnmarshalIPPoolAllocatedIPs(data *string) (inspuripamv1.PoolIPAllocations, error) {
	if data == nil {
		return nil, nil
	}

	var records inspuripamv1.PoolIPAllocations
	if err := json.Unmarshal([]byte(*data), &records); err != nil {
		return nil, err
	}

	return records, nil
}

func MarshalIPPoolAllocatedIPs(records inspuripamv1.PoolIPAllocations) (*string, error) {
	if len(records) == 0 {
		return nil, nil
	}

	v, err := json.Marshal(records)
	if err != nil {
		return nil, err
	}
	data := string(v)

	return &data, nil
}

/*
func UnmarshalSubnetAllocatedIPPools(data *string) (spiderpoolv2beta1.PoolIPPreAllocations, error) {
	if data == nil {
		return nil, nil
	}

	var subnetStatusAllocatedIPPool spiderpoolv2beta1.PoolIPPreAllocations
	err := json.Unmarshal([]byte(*data), &subnetStatusAllocatedIPPool)
	if nil != err {
		return nil, err
	}

	return subnetStatusAllocatedIPPool, nil
}

func MarshalSubnetAllocatedIPPools(preAllocations spiderpoolv2beta1.PoolIPPreAllocations) (*string, error) {
	if len(preAllocations) == 0 {
		return nil, nil
	}

	v, err := json.Marshal(preAllocations)
	if err != nil {
		return nil, err
	}
	data := string(v)

	return &data, nil
}
*/
