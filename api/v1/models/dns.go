// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// DNS k8-ipam DNS
//
// swagger:model DNS
type DNS struct {

	// domain
	Domain string `json:"domain,omitempty"`

	// nameservers
	Nameservers []string `json:"nameservers"`

	// options
	Options []string `json:"options"`

	// search
	Search []string `json:"search"`
}

// Validate validates this DNS
func (m *DNS) Validate(formats strfmt.Registry) error {
	return nil
}

// ContextValidate validates this DNS based on context it is used
func (m *DNS) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *DNS) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *DNS) UnmarshalBinary(b []byte) error {
	var res DNS
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
