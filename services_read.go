package ergani

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

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
