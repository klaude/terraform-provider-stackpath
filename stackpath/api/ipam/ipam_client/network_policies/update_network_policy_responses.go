// Code generated by go-swagger; DO NOT EDIT.

package network_policies

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	"github.com/terraform-providers/terraform-provider-stackpath/stackpath/api/ipam/ipam_models"
)

// UpdateNetworkPolicyReader is a Reader for the UpdateNetworkPolicy structure.
type UpdateNetworkPolicyReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *UpdateNetworkPolicyReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewUpdateNetworkPolicyOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 401:
		result := NewUpdateNetworkPolicyUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewUpdateNetworkPolicyInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		result := NewUpdateNetworkPolicyDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewUpdateNetworkPolicyOK creates a UpdateNetworkPolicyOK with default headers values
func NewUpdateNetworkPolicyOK() *UpdateNetworkPolicyOK {
	return &UpdateNetworkPolicyOK{}
}

/*UpdateNetworkPolicyOK handles this case with default header values.

UpdateNetworkPolicyOK update network policy o k
*/
type UpdateNetworkPolicyOK struct {
	Payload *ipam_models.V1UpdateNetworkPolicyResponse
}

func (o *UpdateNetworkPolicyOK) Error() string {
	return fmt.Sprintf("[PUT /ipam/v1/stacks/{stack_id}/network_policies/{network_policy_id}][%d] updateNetworkPolicyOK  %+v", 200, o.Payload)
}

func (o *UpdateNetworkPolicyOK) GetPayload() *ipam_models.V1UpdateNetworkPolicyResponse {
	return o.Payload
}

func (o *UpdateNetworkPolicyOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(ipam_models.V1UpdateNetworkPolicyResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewUpdateNetworkPolicyUnauthorized creates a UpdateNetworkPolicyUnauthorized with default headers values
func NewUpdateNetworkPolicyUnauthorized() *UpdateNetworkPolicyUnauthorized {
	return &UpdateNetworkPolicyUnauthorized{}
}

/*UpdateNetworkPolicyUnauthorized handles this case with default header values.

Returned when an unauthorized request is attempted.
*/
type UpdateNetworkPolicyUnauthorized struct {
	Payload *ipam_models.APIStatus
}

func (o *UpdateNetworkPolicyUnauthorized) Error() string {
	return fmt.Sprintf("[PUT /ipam/v1/stacks/{stack_id}/network_policies/{network_policy_id}][%d] updateNetworkPolicyUnauthorized  %+v", 401, o.Payload)
}

func (o *UpdateNetworkPolicyUnauthorized) GetPayload() *ipam_models.APIStatus {
	return o.Payload
}

func (o *UpdateNetworkPolicyUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(ipam_models.APIStatus)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewUpdateNetworkPolicyInternalServerError creates a UpdateNetworkPolicyInternalServerError with default headers values
func NewUpdateNetworkPolicyInternalServerError() *UpdateNetworkPolicyInternalServerError {
	return &UpdateNetworkPolicyInternalServerError{}
}

/*UpdateNetworkPolicyInternalServerError handles this case with default header values.

Internal server error.
*/
type UpdateNetworkPolicyInternalServerError struct {
	Payload *ipam_models.APIStatus
}

func (o *UpdateNetworkPolicyInternalServerError) Error() string {
	return fmt.Sprintf("[PUT /ipam/v1/stacks/{stack_id}/network_policies/{network_policy_id}][%d] updateNetworkPolicyInternalServerError  %+v", 500, o.Payload)
}

func (o *UpdateNetworkPolicyInternalServerError) GetPayload() *ipam_models.APIStatus {
	return o.Payload
}

func (o *UpdateNetworkPolicyInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(ipam_models.APIStatus)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewUpdateNetworkPolicyDefault creates a UpdateNetworkPolicyDefault with default headers values
func NewUpdateNetworkPolicyDefault(code int) *UpdateNetworkPolicyDefault {
	return &UpdateNetworkPolicyDefault{
		_statusCode: code,
	}
}

/*UpdateNetworkPolicyDefault handles this case with default header values.

Default error structure.
*/
type UpdateNetworkPolicyDefault struct {
	_statusCode int

	Payload *ipam_models.APIStatus
}

// Code gets the status code for the update network policy default response
func (o *UpdateNetworkPolicyDefault) Code() int {
	return o._statusCode
}

func (o *UpdateNetworkPolicyDefault) Error() string {
	return fmt.Sprintf("[PUT /ipam/v1/stacks/{stack_id}/network_policies/{network_policy_id}][%d] UpdateNetworkPolicy default  %+v", o._statusCode, o.Payload)
}

func (o *UpdateNetworkPolicyDefault) GetPayload() *ipam_models.APIStatus {
	return o.Payload
}

func (o *UpdateNetworkPolicyDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(ipam_models.APIStatus)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}