// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Inspur
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Inspur-Data/ipamwrapper/api/v1/client/ipamwrapper_agent"
	"github.com/Inspur-Data/ipamwrapper/api/v1/models"
	"github.com/Inspur-Data/ipamwrapper/cmd/ipamwrapper-ds/daemonset"
	"github.com/Inspur-Data/ipamwrapper/pkg/config"
	ipTools "github.com/Inspur-Data/ipamwrapper/pkg/ip"
	"github.com/Inspur-Data/ipamwrapper/pkg/logging"
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	cniTypesV1 "github.com/containernetworking/cni/pkg/types/100"
	cniSpecVersion "github.com/containernetworking/cni/pkg/version"
	"github.com/go-openapi/strfmt"
	"net"
	"time"
)

// version means ipamwrapper released version.
var version = "0.1.0"

func main() {
	logging.Debugf("main function will start.....")
	//test()
	skel.PluginMain(cmdAdd, cmdCheck, cmdDel, cniSpecVersion.All, fmt.Sprintf("ipamwrapper version %s", version))
}

func cmdCheck(args *skel.CmdArgs) error {
	return nil
}

func cmdAdd(args *skel.CmdArgs) error {
	logging.Debugf("enter cmdAdd function")
	cniConfig, err := config.ParseConfig(args.StdinData)
	if err != nil {
		return logging.Errorf("parseConfig failed:%v", err)
	}

	podArgs := config.PodArgs{}
	if err = types.LoadArgs(args.Args, &podArgs); nil != err {
		return logging.Errorf("load Pod args failed:%v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	//init a unix client
	//todo add socketPath
	unixAgentAPI, err := daemonset.NewAgentOpenAPIUnixClient(config.CniConfig.IPAM.UnixSocketPath)
	if nil != err {
		return logging.Errorf("failed to create agent client: %v", err)
	}

	//marshal IPAM param to string and assign to post body
	ipamByte, err := json.Marshal(cniConfig.IPAM)
	if err != nil {
		logging.Errorf("marshal IPAM param failed:%v", err)
	}

	ipamStr := string(ipamByte)
	logging.Debugf("IPAM param :%V", ipamStr)
	param := ipamwrapper_agent.NewPostIpamParams().WithContext(ctx).WithIpamAllocArgs(&models.IpamAllocArgs{
		ContainerID:  &args.ContainerID,
		IfName:       &args.IfName,
		NetNamespace: &args.Netns,
		PodName:      (*string)(&podArgs.K8S_POD_NAME),
		PodNamespace: (*string)(&podArgs.K8S_POD_NAMESPACE),
		Ipam:         ipamStr,
	})

	ipamResponse, err := unixAgentAPI.IpamwrapperAgent.PostIpam(param)
	if nil != err {
		logging.Errorf("post ipam alloc failed:%v", err)
		return err
	}

	// check the  request response.
	if err = ipamResponse.Payload.Validate(strfmt.Default); nil != err {
		logging.Errorf("check the response failed:%v", err)
		return err
	}

	//convert the response
	res, err := convertRes(cniConfig.CNIVersion, ipamResponse, args.IfName)
	if err != nil {
		logging.Errorf("convert the response failed:%v", err)
		return err
	}

	return types.PrintResult(res, cniConfig.CNIVersion)

}

func cmdDel(args *skel.CmdArgs) error {
	logging.Debugf("enter cmdDel function")
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
	unixAgentAPI, err := daemonset.NewAgentOpenAPIUnixClient(config.CniConfig.IPAM.UnixSocketPath)
	if nil != err {
		return logging.Errorf("failed to create agent client: %v", err)
	}

	param := ipamwrapper_agent.NewDeleteIpamParams().WithContext(ctx).WithIpamDelArgs(&models.IpamDelArgs{
		ContainerID:  &args.ContainerID,
		IfName:       &args.IfName,
		NetNamespace: args.Netns,
		PodName:      (*string)(&podArgs.K8S_POD_NAME),
		PodNamespace: (*string)(&podArgs.K8S_POD_NAMESPACE),
	})

	_, err = unixAgentAPI.IpamwrapperAgent.DeleteIpam(param)
	if nil != err {
		logging.Errorf("delete ip failed:%v", err)
		return err
	}

	logging.Debugf("delete ip success")
	return nil
}

func convertRes(cniVersion string, response *ipamwrapper_agent.PostIpamOK, interfaceName string) (*cniTypesV1.Result, error) {
	result := &cniTypesV1.Result{
		CNIVersion: cniVersion,
	}

	//add ip to the result
	for _, ip := range response.Payload.Ips {
		if ip != nil {
			if *ip.Nic == interfaceName {
				address, err := ipTools.ParseIP(*ip.Version, *ip.Address, true)
				if err != nil {
					return nil, logging.Errorf("parseIP failed %v", err)
				}
				result.IPs = append(result.IPs, &cniTypesV1.IPConfig{
					Address: *address,
					Gateway: net.ParseIP(ip.Gateway),
				})
			}
		}
	}

	//add route to the result
	for _, route := range response.Payload.Routes {
		if route != nil {
			if *route.IfName == interfaceName {
				_, dst, err := net.ParseCIDR(*route.Dst)
				if err != nil {
					return nil, logging.Errorf("parse CIDR failed %v", err)
				}
				result.Routes = append(result.Routes, &types.Route{
					Dst: *dst,
					GW:  net.ParseIP(*route.Gw),
				})
			}
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
