// Copyright 2022 Authors of Inspur
// SPDX-License-Identifier: Apache-2.0

package daemonset

import (
	"github.com/Inspur-Data/ipamwrapper/api/v1/models"
	IPAMServerAgent "github.com/Inspur-Data/ipamwrapper/api/v1/server/restapi/operations/ipamwrapper_agent"
	"github.com/Inspur-Data/ipamwrapper/pkg/logging"
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
func (g *unixPostIpStruct) Handle(params IPAMServerAgent.PostIpamParams) middleware.Responder {
	logging.Debugf("Enter post handle function,the ipamAllocParams is: %v", *params.IpamAllocArgs)
	if err := params.IpamAllocArgs.Validate(strfmt.Default); err != nil {
		logging.Errorf("post param is invalid: %v", params)
		return IPAMServerAgent.NewPostIpamFailure().WithPayload(models.Error(err.Error()))
	}

	//todo 实现申请IP真正逻辑
	//ctx := context.Background()
	ctx := params.HTTPRequest.Context()
	resp, err := ipamAgent.IPAM.Allocate(ctx, params.IpamAllocArgs)
	if err != nil {
		logging.Errorf("allocate IP failed: %v", err)
		return nil
	}
	return IPAMServerAgent.NewPostIpamOK().WithPayload(resp)
}

type unixDeleteIpStruct struct{}

// Handle implement the logic about release ip,path: /ipam
func (g *unixDeleteIpStruct) Handle(params IPAMServerAgent.DeleteIpamParams) middleware.Responder {
	logging.Debugf("Enter delete handle function")
	if err := params.IpamDelArgs.Validate(strfmt.Default); err != nil {
		return IPAMServerAgent.NewDeleteIpamFailure().WithPayload(models.Error(err.Error()))
	}

	//todo 实现释放IP的具体逻辑
	ctx := params.HTTPRequest.Context()
	err := ipamAgent.IPAM.Delete(ctx, params.IpamDelArgs)
	if err != nil {
		logging.Errorf("delete IP failed:%v", err)
		return IPAMServerAgent.NewDeleteIpamFailure().WithPayload(models.Error(err.Error()))
	}
	return IPAMServerAgent.NewDeleteIpamOK()
}
