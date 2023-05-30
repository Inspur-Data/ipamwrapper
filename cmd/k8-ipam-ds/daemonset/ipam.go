// Copyright 2022 Authors of Inspur
// SPDX-License-Identifier: Apache-2.0

package daemonset

import (
	"github.com/Inspur-Data/k8-ipam/api/v1/models"
	k8IPAMServerAgent "github.com/Inspur-Data/k8-ipam/api/v1/server/restapi/operations/k8_ipam_agent"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
)

// constant
var (
	unixPostIp = &unixPostIpStruct{}

	unixDeleteIp = &unixDeleteIpStruct{}
)

type unixPostIpStruct struct{}

// Handle implement the logic about allocate ip,path: /ipam/ip.
func (g *unixPostIpStruct) Handle(params k8IPAMServerAgent.PostIpamParams) middleware.Responder {
	if err := params.IpamAllocArgs.Validate(strfmt.Default); err != nil {
		return k8IPAMServerAgent.NewPostIpamFailure().WithPayload(models.Error(err.Error()))

	}

	//todo
	//实现申请IP真正逻辑
	/*
		resp, err := agentContext.IPAM.Allocate(ctx, params.IpamAddArgs)
		if err != nil {
			// The count of failures in IP allocations.
			metric.IpamAllocationFailureCounts.Add(ctx, 1)
			gatherIPAMAllocationErrMetric(ctx, err)
			logger.Error(err.Error())

			return daemonset.NewPostIpamIPFailure().WithPayload(models.Error(err.Error()))
		}*/

	//return k8_ipam_agent.NewPostIpamOK().GetPayload()
	resp := models.IpamAllocResponse{}
	return k8IPAMServerAgent.NewPostIpamOK().WithPayload(&resp)
}

type unixDeleteIpStruct struct{}

func (g *unixDeleteIpStruct) Handle(params k8IPAMServerAgent.DeleteIpamParams) middleware.Responder {
	if err := params.IpamDelArgs.Validate(strfmt.Default); err != nil {
		return k8IPAMServerAgent.NewDeleteIpamFailure().WithPayload(models.Error(err.Error()))
	}

	//todo 实现释放IP的具体逻辑

	return k8IPAMServerAgent.NewDeleteIpamOK()
}
