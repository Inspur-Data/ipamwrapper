package v1

import (
	"context"
	"fmt"
	"github.com/Inspur-Data/ipamwrapper/pkg/constant"
	ipamwrapperip "github.com/Inspur-Data/ipamwrapper/pkg/ip"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"strconv"
)

var (
	ipVersionField  *field.Path = field.NewPath("spec").Child("ipVersion")
	cidrField       *field.Path = field.NewPath("spec").Child("cidr")
	ipsField        *field.Path = field.NewPath("spec").Child("ips")
	excludeIPsField *field.Path = field.NewPath("spec").Child("excludeIPs")
	allocateIPField *field.Path = field.NewPath("status").Child("allocatedIPCount")
)

func (r *IPPool) validCreate() field.ErrorList {
	//check ipversion
	if err := r.validIPversion(); err != nil {
		return field.ErrorList{err}
	}

	//check CIDR
	ctx := context.Background()
	if err := r.validCIDR(ctx); err != nil {
		return field.ErrorList{err}
	}

	//check available ip
	if err := r.validAvailableIPs(ctx); err != nil {
		return field.ErrorList{err}
	}

	return nil
}

// validIPversion check the ippool's IP version
func (r *IPPool) validIPversion() *field.Error {
	//check ipversion
	version := r.Spec.IPVersion
	if r.Spec.IPVersion == nil {
		return field.Invalid(
			ipVersionField,
			version,
			"is not generated correctly, 'spec.subnet' may be invalid",
		)
	}

	if *version != constant.IPv4 && *version != constant.IPv6 {
		return field.NotSupported(
			ipVersionField,
			version,
			[]string{strconv.FormatInt(constant.IPv4, 10),
				strconv.FormatInt(constant.IPv6, 10),
			},
		)
	}

	return nil
}

// validCIDR check the ippool's CIDR
func (r *IPPool) validCIDR(ctx context.Context) *field.Error {
	//checkout CIDR
	if err := ipamwrapperip.IsCIDR(*r.Spec.IPVersion, r.Spec.CIDR); err != nil {
		return field.Invalid(
			cidrField,
			r.Spec.CIDR,
			err.Error(),
		)
	}

	var poollist IPPoolList
	if err := ippoolClient.List(ctx, &poollist); err != nil {
		return field.InternalError(cidrField, fmt.Errorf("failed to list IPPools: %v", err))
	}

	for _, pool := range poollist.Items {
		if *pool.Spec.IPVersion == *r.Spec.IPVersion {
			if pool.Name == r.Name {
				return field.InternalError(cidrField, fmt.Errorf("IPPool %s already exists", r.Name))
			}

			if pool.Spec.CIDR == r.Spec.CIDR {
				continue
			}

			overlap, err := ipamwrapperip.IsCIDROverlap(*r.Spec.IPVersion, r.Spec.CIDR, pool.Spec.CIDR)
			if err != nil {
				return field.InternalError(cidrField, fmt.Errorf("failed to judge whether 'spec.CIDR' overlaped: %v", err))
			}

			if overlap {
				return field.Invalid(
					cidrField,
					r.Spec.CIDR,
					fmt.Sprintf("cidr is overlaped with IPPool %s which 'spec.CIDR' is %s", pool.Name, pool.Spec.CIDR),
				)
			}
		}
	}

	return nil
}

// validAvailableIPs check the ippool's available ips
func (r *IPPool) validAvailableIPs(ctx context.Context) *field.Error {
	//validate exclude ips
	for _, excludeIP := range r.Spec.ExcludeIPs {
		valid, err := ipamwrapperip.ContainsIP(*r.Spec.IPVersion, r.Spec.CIDR, excludeIP)
		if err != nil {
			return field.InternalError(excludeIPsField, fmt.Errorf("check contains failed %v", err))
		}

		if !valid {
			return field.Invalid(
				excludeIPsField,
				excludeIP,
				fmt.Sprintf("not pertains to the 'spec.cidr' %s of IPPool", r.Name),
			)
		}
	}

	return nil
}

func (r *IPPool) validDelete() field.ErrorList {
	if *r.Status.AllocatedIPCount != 0 {
		err := field.InternalError(allocateIPField, fmt.Errorf("ippool:%s still has allocated IPs", r.Name))
		return field.ErrorList{err}
	}
	return nil
}
