package ergani

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Service is an interoperability service the employer may run.
//
// The API publishes the list of the services and the parameters they take.
type Service struct {
	// Name is the code of the service.
	Name string

	// Description is the Greek description of the service.
	Description string

	// Parameters are the parameters the service takes.
	Parameters []ServiceParameter
}

// ServiceParameter is a parameter of an interoperability service.
type ServiceParameter struct {
	// Name is the name of the parameter.
	Name string

	// Description is the Greek description of the parameter.
	Description string

	// Required reports whether the service cannot run without the parameter.
	Required bool

	// Type is the type of the parameter. The API uses the names Text, Date,
	// Int, Decimal, ListString, ListInt, ListStringDate, XML and MIME.
	Type string

	// MaxLength bounds the length of the value. It applies to the text
	// parameters.
	MaxLength int
}

// ServiceRequest is a request to run an interoperability service.
type ServiceRequest struct {
	// ServiceCode is the code of the service, as the list of the services
	// reports it.
	ServiceCode string

	// Parameters are the parameters of the service. It is empty when the
	// service takes none.
	Parameters []ServiceParameterValue
}

// ServiceParameterValue is one parameter of a service request.
type ServiceParameterValue struct {
	// Name is the name of the parameter.
	Name string

	// Value is the value of the parameter.
	Value string
}

// serviceRequest is the body of the call that runs a service.
type serviceRequest struct {
	ServiceCode string                  `json:"ServiceCode"`
	Parameters  []serviceParameterValue `json:"Parameters"`
}

// serviceParameterValue is one parameter of a service request.
type serviceParameterValue struct {
	ParameterName  string `json:"ParameterName"`
	ParameterValue string `json:"ParameterValue"`
}

// Services returns the interoperability services the connected employer may
// run, with the parameters they take.
func (c *Client) Services(ctx context.Context) ([]Service, error) {
	data, err := c.call(ctx, http.MethodGet, "/WebServices/ServicesList", nil, nil, true)
	if err != nil {
		return nil, err
	}

	var response []struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Parameters  []struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			IsRequired  bool   `json:"isRequired"`
			Type        string `json:"type"`
			MaxLength   int    `json:"maxLength"`
		} `json:"parameters"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("ergani: failed to decode the services: %w", err)
	}

	services := make([]Service, 0, len(response))
	for _, item := range response {
		service := Service{
			Name:        item.Name,
			Description: item.Description,
		}
		for _, parameter := range item.Parameters {
			service.Parameters = append(service.Parameters, ServiceParameter{
				Name:        parameter.Name,
				Description: parameter.Description,
				Required:    parameter.IsRequired,
				Type:        parameter.Type,
				MaxLength:   parameter.MaxLength,
			})
		}
		services = append(services, service)
	}

	return services, nil
}

// ExecuteService runs an interoperability service and returns its result.
//
// The result is the raw JSON the API answered with, because the shape
// depends on the service.
func (c *Client) ExecuteService(ctx context.Context, req ServiceRequest) (json.RawMessage, error) {
	if req.ServiceCode == "" {
		return nil, &ValidationError{Field: "serviceCode", Reason: "the code of the service is required"}
	}

	parameters := make([]serviceParameterValue, 0, len(req.Parameters))
	for _, parameter := range req.Parameters {
		parameters = append(parameters, serviceParameterValue{
			ParameterName:  parameter.Name,
			ParameterValue: parameter.Value,
		})
	}

	body, err := encode(serviceRequest{ServiceCode: req.ServiceCode, Parameters: parameters})
	if err != nil {
		return nil, err
	}

	data, err := c.call(ctx, http.MethodPost, "/WebServices/ExecuteService", nil, body, true)
	if err != nil {
		return nil, err
	}

	return json.RawMessage(data), nil
}

// CodeList returns the catalogue of the codes of a parameter.
//
// The API publishes the catalogues of the codes it accepts under the
// EX_BASE_03 service. WorkTimeType holds the codes of the working time
// types, WorkCardDelayReason the codes of the late declaration
// justifications, and OvertimeAitiologia the codes of the overtime
// justifications.
//
// The catalogues are the authoritative source of the code tables, and the
// API changes them over time, so they are worth reading before a submission
// rather than relying on the constants of this package.
//
// The result is the raw JSON the API answered with.
func (c *Client) CodeList(ctx context.Context, parameter string) (json.RawMessage, error) {
	if parameter == "" {
		return nil, &ValidationError{Field: "parameter", Reason: "the name of the catalogue is required"}
	}

	return c.ExecuteService(ctx, ServiceRequest{
		ServiceCode: "EX_BASE_03",
		Parameters: []ServiceParameterValue{
			{Name: "Parameter", Value: parameter},
		},
	})
}
