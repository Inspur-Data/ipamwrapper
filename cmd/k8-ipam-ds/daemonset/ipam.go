// Copyright 2022 Authors of Inspur
// SPDX-License-Identifier: Apache-2.0

package daemonset

import (
	"github.com/Inspur-Data/k8-ipam/api/v1/models"
	k8IPAMServerAgent "github.com/Inspur-Data/k8-ipam/api/v1/server/restapi/operations/k8_ipam_agent"
	"github.com/Inspur-Data/k8-ipam/pkg/logging"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
)

// constant
var (
	unixPostIp = &unixPostIpStruct{}

	unixDeleteIp = &unixDeleteIpStruct{}
)

type unixPostIpStruct struct{}

// Handle implement the logic about allocate ip,path: /ipam
func (g *unixPostIpStruct) Handle(params k8IPAMServerAgent.PostIpamParams) middleware.Responder {
	logging.Debugf("Enter post handle function,the ipamAllocParams is: %v", params.IpamAllocArgs)
	if err := params.IpamAllocArgs.Validate(strfmt.Default); err != nil {
		logging.Errorf("post param is invalid: %v", params)
		return k8IPAMServerAgent.NewPostIpamFailure().WithPayload(models.Error(err.Error()))
	}

	//todo 实现申请IP真正逻辑

	resp := models.IpamAllocResponse{}
	return k8IPAMServerAgent.NewPostIpamOK().WithPayload(&resp)
}

type unixDeleteIpStruct struct{}

// Handle implement the logic about release ip,path: /ipam
func (g *unixDeleteIpStruct) Handle(params k8IPAMServerAgent.DeleteIpamParams) middleware.Responder {
	logging.Debugf("Enter delete handle function")
	if err := params.IpamDelArgs.Validate(strfmt.Default); err != nil {
		return k8IPAMServerAgent.NewDeleteIpamFailure().WithPayload(models.Error(err.Error()))
	}

	//todo 实现释放IP的具体逻辑

	return k8IPAMServerAgent.NewDeleteIpamOK()
}
