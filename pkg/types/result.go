// Copyright 2022 Authors of Inspur
// SPDX-License-Identifier: Apache-2.0

package types

import "github.com/Inspur-Data/ipamwrapper/api/v1/models"

type IPVersion = int64

type Vlan = int64

type AllocationResult struct {
	IP           *models.IPConfig
	Routes       []*models.Route
	CleanGateway bool
}

type IPAndUID struct {
	IP  string
	UID string
}

type PoolGroup map[string][]IPAndUID

func (poolgroups *PoolGroup) Pools() []string {
	var pools []string
	for pool := range *poolgroups {
		pools = append(pools, pool)
	}

	return pools
}
