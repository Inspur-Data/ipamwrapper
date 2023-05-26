// Copyright 2022 Authors of Inspur
// SPDX-License-Identifier: Apache-2.0

package ip

import (
	"net"
	"strings"

	"github.com/Inspur-Data/k8-ipam/pkg/constant"
)

func AssembleTotalIPs(ipVersion constant.IPVersion, ipRanges, excludedIPRanges []string) ([]net.IP, error) {
	ips, err := ParseIPRanges(ipVersion, ipRanges)
	if nil != err {
		return nil, err
	}
	excludeIPs, err := ParseIPRanges(ipVersion, excludedIPRanges)
	if nil != err {
		return nil, err
	}
	totalIPs := IPsDiffSet(ips, excludeIPs, false)

	return totalIPs, nil
}

func CIDRToLabelValue(ipVersion constant.IPVersion, subnet string) (string, error) {
	if err := IsCIDR(ipVersion, subnet); err != nil {
		return "", err
	}

	value := strings.Replace(subnet, ".", "-", 3)
	value = strings.Replace(value, ":", "-", 7)
	value = strings.Replace(value, "/", "-", 1)

	return value, nil
}
