package main

import (
	"fmt"
	"github.com/Inspur-Data/ipamwrapper/cmd/router/bin"
	"github.com/containernetworking/cni/pkg/skel"
	cniSpecVersion "github.com/containernetworking/cni/pkg/version"
)

func main() {
	fmt.Sprintf("begin router")
	skel.PluginMain(bin.CmdAdd, nil, bin.CmdDel, cniSpecVersion.All, "Router")
}