// Code generated by go-swagger; DO NOT EDIT.

package ipamwrapper_agent

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/Inspur-Data/ipamwrapper/api/v1/models"
)

// DeleteIpamOKCode is the HTTP code returned for type DeleteIpamOK
const DeleteIpamOKCode int = 200

/*
DeleteIpamOK Success

swagger:response deleteIpamOK
*/
type DeleteIpamOK struct {
}

// NewDeleteIpamOK creates DeleteIpamOK with default headers values
func NewDeleteIpamOK() *DeleteIpamOK {

	return &DeleteIpamOK{}
}

// WriteResponse to the client
func (o *DeleteIpamOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(200)
}

// DeleteIpamFailureCode is the HTTP code returned for type DeleteIpamFailure
const DeleteIpamFailureCode int = 500

/*
DeleteIpamFailure Addresses release failure

swagger:response deleteIpamFailure
*/
type DeleteIpamFailure struct {

	/*
	  In: Body
	*/
	Payload models.Error `json:"body,omitempty"`
}

// NewDeleteIpamFailure creates DeleteIpamFailure with default headers values
func NewDeleteIpamFailure() *DeleteIpamFailure {

	return &DeleteIpamFailure{}
}

// WithPayload adds the payload to the delete ipam failure response
func (o *DeleteIpamFailure) WithPayload(payload models.Error) *DeleteIpamFailure {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the delete ipam failure response
func (o *DeleteIpamFailure) SetPayload(payload models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *DeleteIpamFailure) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(500)
	payload := o.Payload
	if err := producer.Produce(rw, payload); err != nil {
		panic(err) // let the recovery middleware deal with this
	}
}
