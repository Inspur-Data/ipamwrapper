// Copyright 2022 Authors of Inspur
// SPDX-License-Identifier: Apache-2.0

package ip

import "errors"

var (
	ErrInvalidIPVersion     = errors.New("invalid IP version")
	ErrInvalidIPRangeFormat = errors.New("invalid IP range format")
	ErrInvalidIPFormat      = errors.New("invalid IP format")
	ErrInvalidCIDRFormat    = errors.New("invalid CIDR format")
	ErrInvalidRouteFormat   = errors.New("invalid route format")
	ErrInvalidIP            = errors.New("invalid IP")
)
