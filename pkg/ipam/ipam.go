// Copyright 2022 Authors of Inspur
// SPDX-License-Identifier: Apache-2.0

package ipam

import (
	"context"
	"github.com/Inspur-Data/k8-ipam/api/v1/models"
	"github.com/Inspur-Data/k8-ipam/pkg/logging"
	"github.com/Inspur-Data/k8-ipam/pkg/manager/podmanager"
)

type IPAM interface {
	Allocate(ctx context.Context, addArgs *models.IpamAllocArgs) (*models.IpamAllocResponse, error)
	Delete(ctx context.Context, delArgs *models.IpamDelArgs) error
	Start(ctx context.Context) error
}

type ipam struct {
	config     IPAMConfig
	podManager podmanager.PodManager
}

// NewIPAM init a new IPAM instance
func NewIPAM(config IPAMConfig, podManager podmanager.PodManager) (IPAM, error) {
	if podManager == nil {
		return nil, logging.Errorf("podManager is nil")
	}
	return &ipam{
		podManager: podManager,
		config:     config,
	}, nil
}

// Allocate will allocate an IP with the given param
func (i *ipam) Allocate(ctx context.Context, addArgs *models.IpamAllocArgs) (*models.IpamAllocResponse, error) {
	return nil, nil
}

// Delete release the ip with the given param
func (i *ipam) Delete(ctx context.Context, delArgs *models.IpamDelArgs) error {
	return nil
}

// Start will start the IPAM instance
func (i *ipam) Start(ctx context.Context) error {
	return nil
}
