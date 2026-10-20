package ergani

import (
	"context"
	"encoding/json"
	"fmt"
)

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
