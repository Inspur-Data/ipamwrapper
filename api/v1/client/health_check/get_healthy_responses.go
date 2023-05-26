// Code generated by go-swagger; DO NOT EDIT.

package health_check

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
)

// GetHealthyReader is a Reader for the GetHealthy structure.
type GetHealthyReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetHealthyReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewGetHealthyOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 500:
		result := NewGetHealthyInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewGetHealthyOK creates a GetHealthyOK with default headers values
func NewGetHealthyOK() *GetHealthyOK {
	return &GetHealthyOK{}
}

/*
GetHealthyOK describes a response with status code 200, with default header values.

Success
*/
type GetHealthyOK struct {
}

// IsSuccess returns true when this get healthy o k response has a 2xx status code
func (o *GetHealthyOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this get healthy o k response has a 3xx status code
func (o *GetHealthyOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get healthy o k response has a 4xx status code
func (o *GetHealthyOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this get healthy o k response has a 5xx status code
func (o *GetHealthyOK) IsServerError() bool {
	return false
}

// IsCode returns true when this get healthy o k response a status code equal to that given
func (o *GetHealthyOK) IsCode(code int) bool {
	return code == 200
}

// Code gets the status code for the get healthy o k response
func (o *GetHealthyOK) Code() int {
	return 200
}

func (o *GetHealthyOK) Error() string {
	return fmt.Sprintf("[GET /healthy][%d] getHealthyOK ", 200)
}

func (o *GetHealthyOK) String() string {
	return fmt.Sprintf("[GET /healthy][%d] getHealthyOK ", 200)
}

func (o *GetHealthyOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewGetHealthyInternalServerError creates a GetHealthyInternalServerError with default headers values
func NewGetHealthyInternalServerError() *GetHealthyInternalServerError {
	return &GetHealthyInternalServerError{}
}

/*
GetHealthyInternalServerError describes a response with status code 500, with default header values.

Failed
*/
type GetHealthyInternalServerError struct {
}

// IsSuccess returns true when this get healthy internal server error response has a 2xx status code
func (o *GetHealthyInternalServerError) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this get healthy internal server error response has a 3xx status code
func (o *GetHealthyInternalServerError) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get healthy internal server error response has a 4xx status code
func (o *GetHealthyInternalServerError) IsClientError() bool {
	return false
}

// IsServerError returns true when this get healthy internal server error response has a 5xx status code
func (o *GetHealthyInternalServerError) IsServerError() bool {
	return true
}

// IsCode returns true when this get healthy internal server error response a status code equal to that given
func (o *GetHealthyInternalServerError) IsCode(code int) bool {
	return code == 500
}

// Code gets the status code for the get healthy internal server error response
func (o *GetHealthyInternalServerError) Code() int {
	return 500
}

func (o *GetHealthyInternalServerError) Error() string {
	return fmt.Sprintf("[GET /healthy][%d] getHealthyInternalServerError ", 500)
}

func (o *GetHealthyInternalServerError) String() string {
	return fmt.Sprintf("[GET /healthy][%d] getHealthyInternalServerError ", 500)
}

func (o *GetHealthyInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}
