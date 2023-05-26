// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// IpamAllocResponse Alloc IP information
//
// swagger:model IpamAllocResponse
type IpamAllocResponse struct {

	// dns
	DNS *DNS `json:"dns,omitempty"`

	// ip
	// Required: true
	IP *IPConfig `json:"ip"`

	// route
	Route *Route `json:"route,omitempty"`
}

// Validate validates this ipam alloc response
func (m *IpamAllocResponse) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateDNS(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateIP(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateRoute(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *IpamAllocResponse) validateDNS(formats strfmt.Registry) error {
	if swag.IsZero(m.DNS) { // not required
		return nil
	}

	if m.DNS != nil {
		if err := m.DNS.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("dns")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("dns")
			}
			return err
		}
	}

	return nil
}

func (m *IpamAllocResponse) validateIP(formats strfmt.Registry) error {

	if err := validate.Required("ip", "body", m.IP); err != nil {
		return err
	}

	if m.IP != nil {
		if err := m.IP.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("ip")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("ip")
			}
			return err
		}
	}

	return nil
}

func (m *IpamAllocResponse) validateRoute(formats strfmt.Registry) error {
	if swag.IsZero(m.Route) { // not required
		return nil
	}

	if m.Route != nil {
		if err := m.Route.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("route")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("route")
			}
			return err
		}
	}

	return nil
}

// ContextValidate validate this ipam alloc response based on the context it is used
func (m *IpamAllocResponse) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	var res []error

	if err := m.contextValidateDNS(ctx, formats); err != nil {
		res = append(res, err)
	}

	if err := m.contextValidateIP(ctx, formats); err != nil {
		res = append(res, err)
	}

	if err := m.contextValidateRoute(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *IpamAllocResponse) contextValidateDNS(ctx context.Context, formats strfmt.Registry) error {

	if m.DNS != nil {
		if err := m.DNS.ContextValidate(ctx, formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("dns")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("dns")
			}
			return err
		}
	}

	return nil
}

func (m *IpamAllocResponse) contextValidateIP(ctx context.Context, formats strfmt.Registry) error {

	if m.IP != nil {
		if err := m.IP.ContextValidate(ctx, formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("ip")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("ip")
			}
			return err
		}
	}

	return nil
}

func (m *IpamAllocResponse) contextValidateRoute(ctx context.Context, formats strfmt.Registry) error {

	if m.Route != nil {
		if err := m.Route.ContextValidate(ctx, formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("route")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("route")
			}
			return err
		}
	}

	return nil
}

// MarshalBinary interface implementation
func (m *IpamAllocResponse) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *IpamAllocResponse) UnmarshalBinary(b []byte) error {
	var res IpamAllocResponse
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
