package v1

import (
	"github.com/Inspur-Data/ipamwrapper/pkg/constant"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"strconv"
)

var (
	ipVersionField  *field.Path = field.NewPath("spec").Child("ipVersion")
	subnetField     *field.Path = field.NewPath("spec").Child("subnet")
	ipsField        *field.Path = field.NewPath("spec").Child("ips")
	excludeIPsField *field.Path = field.NewPath("spec").Child("excludeIPs")
)

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
