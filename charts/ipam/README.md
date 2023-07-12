# ipamwrapper

## Introduction

The ipamwrapper is an IP Address Management (IPAM) CNI plugin that assigns IP addresses for kubernetes clusters.

Any Container Network Interface (CNI) plugin supporting third-party IPAM plugins can use the ipamwrapper.

## Why ipamwrapper

Most overlay CNIs, like
[Cilium](https://github.com/cilium/cilium)
and [Calico](https://github.com/projectcalico/calico),
have a good implementation of IPAM, so the ipamwrapper is not intentionally designed for these cases, but maybe integrated with them.

The ipamwrapper is intentionally designed to use with underlay network, where administrators can accurately manage each IP.

Currently, in the community, the IPAM plugins such as [whereabout](https://github.com/k8snetworkplumbingwg/whereabouts), [kube-ipam](https://github.com/cloudnativer/kube-ipam),
[static](https://github.com/containernetworking/plugins/tree/main/plugins/ipam/static),
[dhcp](https://github.com/containernetworking/plugins/tree/main/plugins/ipam/dhcp), and [host-local](https://github.com/containernetworking/plugins/tree/main/plugins/ipam/host-local),
few of them could help solve complex underlay-network issues, so we decide to develop the ipamwrapper.

BTW, there are also some CNI plugins that could work on the underlay mode, such as [kube-ovn](https://github.com/kubeovn/kube-ovn) and [coil](https://github.com/cybozu-go/coil).
But the ipamwrapper provides lots of different features, you could see [Features](#features) for details.

