package ergani

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
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

// SubmitDailyScheduleVariant submits a daily schedule declaration of any
// variant.
//
// The variant selects the type of the declaration: the daily one, its
// retrospective variant, or the one of the drivers. The payload is the same
// for the three of them.
func (c *Client) SubmitDailyScheduleVariant(ctx context.Context, variant ScheduleVariant, declarations []DailyScheduleDeclaration) ([]Submission, error) {
	switch variant {
	case ScheduleDaily, ScheduleDailyRetrospective, ScheduleDailyDrivers:
	default:
		return nil, &ValidationError{
			Field:  "variant",
			Value:  variant,
			Reason: "the variant is not a daily schedule declaration",
		}
	}

	payload, err := buildDailySchedules(declarations)
	if err != nil {
		return nil, err
	}

	return c.submit(ctx, DocumentCode(variant), payload)
}

// SubmitOvertimeScheduleVariant submits an overtime working time declaration
// of any variant.
//
// The variant selects the type of the declaration: the overtime itself, its
// retrospective variant, or the one of the drivers. The payload is the same
// for the three of them.
func (c *Client) SubmitOvertimeScheduleVariant(ctx context.Context, variant ScheduleVariant, declarations []OvertimeScheduleDeclaration) ([]Submission, error) {
	switch variant {
	case ScheduleOvertime, ScheduleOvertimeRetrospective, ScheduleOvertimeDrivers:
	default:
		return nil, &ValidationError{
			Field:  "variant",
			Value:  variant,
			Reason: "the variant is not an overtime declaration",
		}
	}

	payload, err := buildOvertimeSchedules(declarations)
	if err != nil {
		return nil, err
	}

	return c.submit(ctx, DocumentCode(variant), payload)
}

// Employer is the master data of the employer.
//
// It is what the service EX_BASE_01 answers with.
type Employer struct {
	// ID is the identifier of the employer.
	ID string `json:"Id"`

	// TaxID is the tax identification number of the employer.
	TaxID string `json:"Afm"`

	// Name is the name of the employer.
	Name string `json:"Eponimia"`

	// TradeName is the commercial name of the employer.
	TradeName string `json:"DiakritikosTitlos"`

	// AME is the registration number of the employer.
	AME string `json:"Ame"`

	// InCardSector reports whether the employer belongs to the sector of the
	// digital work card.
	InCardSector bool `json:"IsInCardSector"`
}

// Branch is the master data of one branch of the employer.
//
// It is what the service EX_BASE_02 answers with.
type Branch struct {
	// BranchNumber is the number of the branch.
	BranchNumber string `json:"Aa"`

	// Address is the address of the branch.
	Address string `json:"Address"`

	// SEPEServiceCode is the code of the labour inspection service of the
	// branch.
	SEPEServiceCode string `json:"YpiresiaSepe"`

	// OAEEDServiceCode is the code of the employment service of the branch.
	OAEEDServiceCode string `json:"YpiresiaOaed"`

	// ActivityCode is the activity code of the branch.
	ActivityCode string `json:"Kad"`

	// MunicipalCode is the municipal code of the branch.
	MunicipalCode string `json:"Kallikratis"`

	// StatusDescription is the description of the status of the branch.
	StatusDescription string `json:"StatusDescription"`
}

// EmployerDetails reads the master data of the employer.
func (c *Client) EmployerDetails(ctx context.Context) (*Employer, error) {
	result, err := c.ExecuteService(ctx, ServiceRequest{ServiceCode: ServiceEmployerDetails})
	if err != nil {
		return nil, err
	}

	var response struct {
		EXBASE01 struct {
			Employer Employer `json:"Ergodotis"`
		} `json:"EX_BASE_01"`
	}
	if err := json.Unmarshal(result, &response); err != nil {
		return nil, fmt.Errorf("ergani: failed to decode the details of the employer: %w", err)
	}

	return &response.EXBASE01.Employer, nil
}

// BranchDetails reads the master data of the branches of the employer.
func (c *Client) BranchDetails(ctx context.Context) ([]Branch, error) {
	result, err := c.ExecuteService(ctx, ServiceRequest{ServiceCode: ServiceBranchDetails})
	if err != nil {
		return nil, err
	}

	var response struct {
		EXBASE02 struct {
			Branches []Branch `json:"Pararthma"`
		} `json:"EX_BASE_02"`
	}
	if err := json.Unmarshal(result, &response); err != nil {
		return nil, fmt.Errorf("ergani: failed to decode the details of the branches: %w", err)
	}

	return response.EXBASE02.Branches, nil
}

// The interoperability services that read back what the platform recorded.
//
// The service codes are the ones the API publishes in the list of the
// services.
const (
	// ServiceEmployerDetails reads the master data of the employer.
	ServiceEmployerDetails = "EX_BASE_01"

	// ServiceBranchDetails reads the master data of the branches.
	ServiceBranchDetails = "EX_BASE_02"

	// ServiceCodeLists reads the catalogues of the codes the API accepts.
	ServiceCodeLists = "EX_BASE_03"

	// ServiceMonthlyStatus reads the monthly employment state, the counters
	// the labour inspector sees.
	ServiceMonthlyStatus = "EX_BASE_04"

	// ServiceCurrentWorkforce reads the employees of the employer.
	ServiceCurrentWorkforce = "EX_BASE_05"

	// ServiceAcceptanceStatus reads the acceptance of the essential terms.
	ServiceAcceptanceStatus = "EX_BASE_06"

	// ServiceWorkingDiary reads the diary of the actual hours.
	ServiceWorkingDiary = "EX_BASE_07"

	// ServiceWorkingTimeStatus reads the digital working time status.
	ServiceWorkingTimeStatus = "EX_BASE_08"
)

// WorkingEntry is one day of the diary of the actual hours.
//
// It is what the service EX_BASE_07 answers with.
type WorkingEntry struct {
	// BranchNumber is the number of the branch of the employer.
	BranchNumber string `json:"Aa"`

	// EmployeeTaxID is the tax identification number of the employee.
	EmployeeTaxID string `json:"Afm"`

	// Date is the day the entry covers.
	Date string `json:"Date"`

	// FromTime is the moment the employee started.
	FromTime string `json:"HourFrom"`

	// ToTime is the moment the employee stopped.
	ToTime string `json:"HourTo"`

	// EndDateDiffers reports whether the day of the end differs from the
	// day of the start.
	EndDateDiffers bool `json:"IsEndDateDifferentThanDate"`
}

// WorkingTimeEntry is the digital working time status of one employee on one
// day.
//
// It is what the service EX_BASE_08 answers with.
type WorkingTimeEntry struct {
	// BranchNumber is the number of the branch of the employer.
	BranchNumber string `json:"aa"`

	// EmployeeTaxID is the tax identification number of the employee.
	EmployeeTaxID string `json:"Afm"`

	// Date is the day the entry covers.
	Date string `json:"Date"`

	// Type is the kind of the working period.
	Type string `json:"Type"`

	// FromTime is the moment the period started.
	FromTime string `json:"HourFrom"`

	// ToTime is the moment the period ended.
	ToTime string `json:"HourTo"`

	// Extra reports whether the period is an extra one.
	Extra bool `json:"Extra"`

	// BreakMinutes are the minutes of the break.
	BreakMinutes int `json:"BreakMinutes"`

	// BreakInWork reports whether the break is inside the working period.
	BreakInWork bool `json:"BreakInWork"`
}

// WorkingDiary reads the diary of the actual hours of the employees.
//
// The days before the current month are the ones the platform answers with.
func (c *Client) WorkingDiary(ctx context.Context, branch string, date time.Time) ([]WorkingEntry, error) {
	result, err := c.ExecuteService(ctx, ServiceRequest{
		ServiceCode: ServiceWorkingDiary,
		Parameters: []ServiceParameterValue{
			{Name: "PararthmaAa", Value: branch},
			{Name: "Date", Value: date.Format("02/01/2006")},
		},
	})
	if err != nil {
		return nil, err
	}

	var response struct {
		EXBASE07 struct {
			RealWorking []WorkingEntry `json:"RealWorking"`
		} `json:"EX_BASE_07"`
	}
	if err := json.Unmarshal(result, &response); err != nil {
		return nil, fmt.Errorf("ergani: failed to decode the working diary: %w", err)
	}

	return response.EXBASE07.RealWorking, nil
}

// WorkingTimeStatus reads the digital working time status of the employees.
//
// The days before the current month are the ones the platform answers with.
func (c *Client) WorkingTimeStatus(ctx context.Context, branch string, date time.Time) ([]WorkingTimeEntry, error) {
	result, err := c.ExecuteService(ctx, ServiceRequest{
		ServiceCode: ServiceWorkingTimeStatus,
		Parameters: []ServiceParameterValue{
			{Name: "PararthmaAa", Value: branch},
			{Name: "Date", Value: date.Format("02/01/2006")},
		},
	})
	if err != nil {
		return nil, err
	}

	var response struct {
		EXBASE08 struct {
			Working []WorkingTimeEntry `json:"Working"`
		} `json:"EX_BASE_08"`
	}
	if err := json.Unmarshal(result, &response); err != nil {
		return nil, fmt.Errorf("ergani: failed to decode the working time status: %w", err)
	}

	return response.EXBASE08.Working, nil
}

// CurrentWorkforce reads the employees of the employer.
//
// The tax identification number of a single employee narrows the answer. The
// raw payload is kept, because the service answers with more than eighty
// fields and the platform adds more over time.
func (c *Client) CurrentWorkforce(ctx context.Context, taxID string) (json.RawMessage, error) {
	request := ServiceRequest{ServiceCode: ServiceCurrentWorkforce}
	if taxID != "" {
		request.Parameters = append(request.Parameters, ServiceParameterValue{Name: "afm", Value: taxID})
	}

	return c.ExecuteService(ctx, request)
}

// MonthlyStatus reads the monthly employment state, the counters the labour
// inspector sees.
//
// The raw payload is kept, because the service answers with more than sixty
// counters and the platform adds more over time.
func (c *Client) MonthlyStatus(ctx context.Context, year, month int) (json.RawMessage, error) {
	if month < 1 || month > 12 {
		return nil, &ValidationError{Field: "month", Reason: "the month must be between 1 and 12"}
	}

	return c.ExecuteService(ctx, ServiceRequest{
		ServiceCode: ServiceMonthlyStatus,
		Parameters: []ServiceParameterValue{
			{Name: "ReportYear", Value: fmt.Sprintf("%04d", year)},
			{Name: "ReportMonth", Value: fmt.Sprintf("%02d", month)},
		},
	})
}
