// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Inspur
package main

import (
	"context"
	"fmt"
	"github.com/Inspur-Data/k8-ipam/api/v1/client/k8_ipam_agent"
	"github.com/Inspur-Data/k8-ipam/api/v1/models"
	"github.com/Inspur-Data/k8-ipam/cmd/k8-ipam-ds/daemonset"
	"github.com/Inspur-Data/k8-ipam/pkg/config"
	ipTools "github.com/Inspur-Data/k8-ipam/pkg/ip"
	"github.com/Inspur-Data/k8-ipam/pkg/logging"
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	cniTypesV1 "github.com/containernetworking/cni/pkg/types/100"
	cniSpecVersion "github.com/containernetworking/cni/pkg/version"
	"github.com/go-openapi/strfmt"
	"net"
	"time"
)

// version means k8-ipam released version.
var version string

func main() {
	logging.Debugf("main function will start.....")
	skel.PluginMain(cmdAdd, cmdCheck, cmdDel, cniSpecVersion.All, fmt.Sprintf("k8-ipam version %s", version))
}

func cmdCheck(args *skel.CmdArgs) error {
	return nil
}

func cmdAdd(args *skel.CmdArgs) error {
	logging.Debugf("Enter cmdAdd function")
	cniConfig, err := config.ParseConfig(args.StdinData)
	if err != nil {
		return logging.Errorf("ParseConfig failed:%v", err)
	}

	cniConfig.IPAM.Type = ""
	podArgs := config.PodArgs{}
	if err = types.LoadArgs(args.Args, &podArgs); nil != err {
		return logging.Errorf("Load Pod args failed:%v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	//init a unix client
	//todo unixSocketPath
	unixAgentAPI, err := daemonset.NewAgentOpenAPIUnixClient("")
	if nil != err {
		return logging.Errorf("failed to create agent client: %v", err)
	}

	param := k8_ipam_agent.NewPostIpamParams().WithContext(ctx).WithIpamAllocArgs(&models.IpamAllocArgs{
		ContainerID:  &args.ContainerID,
		IfName:       &args.IfName,
		NetNamespace: &args.Netns,
		PodName:      (*string)(&podArgs.K8S_POD_NAME),
		PodNamespace: (*string)(&podArgs.K8S_POD_NAMESPACE),
	})

	ipamResponse, err := unixAgentAPI.K8IpamAgent.PostIpam(param)
	if nil != err {
		logging.Errorf("Post ipam alloc failed:%v", err)
		return err
	}

	// check the  request response.
	if err = ipamResponse.Payload.Validate(strfmt.Default); nil != err {
		logging.Errorf("Check the response failed:%v", err)
		return err
	}

	//convert the response
	res, err := convertRes(cniConfig.CNIVersion, ipamResponse, args.IfName)
	if err != nil {
		logging.Errorf("Convert the response failed:%v", err)
		return err
	}

	return types.PrintResult(res, cniConfig.CNIVersion)

}

func cmdDel(args *skel.CmdArgs) error {
	logging.Debugf("Enter cmdDel function")
	cniConfig, err := config.ParseConfig(args.StdinData)
	if err != nil {
		return logging.Errorf("ParseConfig failed")
	}

	cniConfig.IPAM.Type = ""
	podArgs := config.PodArgs{}
	if err = types.LoadArgs(args.Args, &podArgs); nil != err {
		return logging.Errorf("Load Pod args failed:%v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	//init a unix client
	//todo unixSockerPath
	unixAgentAPI, err := daemonset.NewAgentOpenAPIUnixClient("")
	if nil != err {
		return logging.Errorf("failed to create agent client: %v", err)
	}

	param := k8_ipam_agent.NewDeleteIpamParams().WithContext(ctx).WithIpamDelArgs(&models.IpamDelArgs{
		ContainerID:  &args.ContainerID,
		IfName:       &args.IfName,
		NetNamespace: args.Netns,
		PodName:      (*string)(&podArgs.K8S_POD_NAME),
		PodNamespace: (*string)(&podArgs.K8S_POD_NAMESPACE),
	})

	_, err = unixAgentAPI.K8IpamAgent.DeleteIpam(param)
	if nil != err {
		logging.Errorf("Delete ip failed:%v", err)
		return err
	}

	logging.Debugf("Delete IP success")
	return nil
}

func convertRes(cniVersion string, response *k8_ipam_agent.PostIpamOK, interfaceName string) (*cniTypesV1.Result, error) {
	result := &cniTypesV1.Result{
		CNIVersion: cniVersion,
	}

	//add ip to the result
	ip := response.Payload.IP
	if ip != nil {
		if *ip.Nic == interfaceName {
			address, err := ipTools.ParseIP(*ip.Version, *ip.Address, true)
			if err != nil {
				return nil, logging.Errorf("ParseIP failed %v", err)
			}
			result.IPs = append(result.IPs, &cniTypesV1.IPConfig{
				Address: *address,
				Gateway: net.ParseIP(ip.Gateway),
			})
		}
	}

	//add route to the result
	route := response.Payload.Route
	if route != nil {
		if *route.IfName == interfaceName {
			_, dst, err := net.ParseCIDR(*route.Dst)
			if err != nil {
				return nil, logging.Errorf("Parse CIDR failed %v", err)
			}
			result.Routes = append(result.Routes, &types.Route{
				Dst: *dst,
				GW:  net.ParseIP(*route.Gw),
			})
		}
	}

	//add dns to the result
	dns := response.Payload.DNS
	if dns != nil {
		result.DNS = types.DNS{
			Nameservers: response.Payload.DNS.Nameservers,
			Domain:      response.Payload.DNS.Domain,
			Search:      response.Payload.DNS.Search,
			Options:     response.Payload.DNS.Options,
		}
	}
	return result, nil
}
