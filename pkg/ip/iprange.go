// Copyright 2022 Authors of Inspur
// SPDX-License-Identifier: Apache-2.0

package ip

import (
	"bytes"
	"fmt"
	"github.com/Inspur-Data/ipamwrapper/pkg/logging"
	"net"
	"sort"
	"strings"

	"github.com/asaskevich/govalidator"

	"github.com/Inspur-Data/ipamwrapper/pkg/constant"
)

// MergeIPRanges merges dispersed IP ranges.
// For example, transport [172.18.40.1-172.18.40.3, 172.18.40.2-172.18.40.5]
// to [172.18.40.1-172.18.40.5]. The overlapping part of two IP ranges will
// be ignored.
func MergeIPRanges(version constant.IPVersion, ipRanges []string) ([]string, error) {
	ips, err := ParseIPRanges(version, ipRanges)
	if err != nil {
		return nil, err
	}

	return ConvertIPsToIPRanges(version, ips)
}

// ParseIPRanges parses IP ranges as a IP address slices of the specified
// IP version.
func ParseIPRanges(version constant.IPVersion, ipRanges []string) ([]net.IP, error) {
	var sum []net.IP
	for _, r := range ipRanges {
		ips, err := ParseIPRange(version, r)
		if err != nil {
			return nil, err
		}
		sum = append(sum, ips...)
	}

	return sum, nil
}

// ParseIPRange parses IP range as an IP address slices of the specified
// IP version.
func ParseIPRange(version constant.IPVersion, ipRange string) ([]net.IP, error) {
	if err := IsIPRange(version, ipRange); err != nil {
		return nil, err
	}

	arr := strings.Split(ipRange, "-")
	n := len(arr)
	var ips []net.IP
	if n == 1 {
		ips = append(ips, net.ParseIP(arr[0]))
	}

	if n == 2 {
		cur := net.ParseIP(arr[0])
		end := net.ParseIP(arr[1])
		for Cmp(cur, end) <= 0 {
			ips = append(ips, cur)
			cur = NextIP(cur)
		}
	}

	return ips, nil
}

// ConvertIPsToIPRanges converts the IP address slices of the specified
// IP version into a group of distinct, sorted and merged IP ranges.
func ConvertIPsToIPRanges(version constant.IPVersion, ips []net.IP) ([]string, error) {
	if err := IsIPVersion(version); err != nil {
		return nil, err
	}

	set := map[string]struct{}{}
	for _, ip := range ips {
		if (version == constant.IPv4 && ip.To4() == nil) ||
			(version == constant.IPv6 && ip.To4() != nil) {
			return nil, logging.Errorf("%wv%d IP '%s'", ErrInvalidIP, version, ip.String())
		}
		set[ip.String()] = struct{}{}
	}

	ips = ips[0:0]
	for v := range set {
		ips = append(ips, net.ParseIP(v))
	}

	sort.Slice(ips, func(i, j int) bool {
		return bytes.Compare(ips[i].To16(), ips[j].To16()) < 0
	})

	var ipRanges []string
	var start, end int
	for {
		if start == len(ips) {
			break
		}

		if end+1 < len(ips) && ips[end+1].Equal(NextIP(ips[end])) {
			end++
			continue
		}

		if start == end {
			ipRanges = append(ipRanges, ips[start].String())
		} else {
			ipRanges = append(ipRanges, fmt.Sprintf("%s-%s", ips[start], ips[end]))
		}

		start = end + 1
		end = start
	}

	return ipRanges, nil
}

// ContainsIPRange reports whether the subnet parsed from the subnet string
// includes the IP address slices parsed from the IP range. Both must belong
// to the same IP version.
func ContainsIPRange(version constant.IPVersion, subnet string, ipRange string) (bool, error) {
	ipNet, err := ParseCIDR(version, subnet)
	if err != nil {
		return false, err
	}
	ips, err := ParseIPRange(version, ipRange)
	if err != nil {
		return false, err
	}

	n := len(ips)
	if n == 1 {
		return ipNet.Contains(ips[0]), nil
	}

	return ipNet.Contains(ips[0]) && ipNet.Contains(ips[n-1]), nil
}

// IsIPRangeOverlap reports whether the IP address slices of specific IP
// version parsed from two IP ranges overlap.
func IsIPRangeOverlap(version constant.IPVersion, ipRange1, ipRange2 string) (bool, error) {
	if err := IsIPVersion(version); err != nil {
		return false, err
	}
	if err := IsIPRange(version, ipRange1); err != nil {
		return false, err
	}
	if err := IsIPRange(version, ipRange2); err != nil {
		return false, err
	}

	// Ignore the error returned here. The format of the IP range has been
	// verified in IsIPRange above.
	ips1, _ := ParseIPRange(version, ipRange1)
	ips2, _ := ParseIPRange(version, ipRange2)
	if len(ips1) > len(IPsDiffSet(ips1, ips2, false)) {
		return true, nil
	}

	return false, nil
}

// IsIPRange reports whether ipRange string is a valid IP range. An IP
// range can be a single IP address in the style of '172.18.40.0', or
// an address range in the form of '172.18.40.0-172.18.40.10'.
// The following formats are invalid:
// "172.18.40.0 - 172.18.40.10": there can be no space between two IP
// addresses.
// "172.18.40.1-2001:db8:a0b:12f0::1": invalid combination of IPv4 and
// IPv6.
// "172.18.40.10-172.18.40.1": the IP range must be ordered.
func IsIPRange(version constant.IPVersion, ipRange string) error {
	if err := IsIPVersion(version); err != nil {
		return err
	}

	if (version == constant.IPv4 && !IsIPv4IPRange(ipRange)) ||
		(version == constant.IPv6 && !IsIPv6IPRange(ipRange)) {
		return logging.Errorf("%w in IPv%d '%s'", ErrInvalidIPRangeFormat, version, ipRange)
	}

	return nil
}

// IsIPv4IPRange reports whether ipRange string is a valid IPv4 range.
// See IsIPRange for more description of IP range.
func IsIPv4IPRange(ipRange string) bool {
	ips := strings.Split(ipRange, "-")
	n := len(ips)
	if n > 2 {
		return false
	}

	if n == 1 {
		return govalidator.IsIPv4(ips[0])
	}

	if n == 2 {
		if !govalidator.IsIPv4(ips[0]) || !govalidator.IsIPv4(ips[1]) {
			return false
		}
		if Cmp(net.ParseIP(ips[0]), net.ParseIP(ips[1])) == 1 {
			return false
		}
	}

	return true
}

// IsIPv6IPRange reports whether ipRange string is a valid IPv6 range.
// See IsIPRange for more description of IP range.
func IsIPv6IPRange(ipRange string) bool {
	ips := strings.Split(ipRange, "-")
	n := len(ips)
	if n > 2 {
		return false
	}

	if n == 1 {
		return govalidator.IsIPv6(ips[0])
	}

	if n == 2 {
		if !govalidator.IsIPv6(ips[0]) || !govalidator.IsIPv6(ips[1]) {
			return false
		}
		if Cmp(net.ParseIP(ips[0]), net.ParseIP(ips[1])) == 1 {
			return false
		}
	}

	return true
}
