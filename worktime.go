package ergani

import (
	"context"
	"fmt"
	"github.com/takispanag/ergani-go-sdk/internal/wire"
	"time"
)

// ScheduleVariant is a type of the working time organisation declaration.
//
// The value is the code of the declaration type, as the API expects it in
// the path of the submission.
type ScheduleVariant string

// The working time organisation declarations.
const (
	// ScheduleDaily is the declaration of a working time that varies from day
	// to day.
	ScheduleDaily ScheduleVariant = "WTODaily"

	// ScheduleDailyRetrospective is the retrospective variant of the variable
	// daily working time declaration. It is filed by the branches that are
	// exempt from the duty of advance notification.
	ScheduleDailyRetrospective ScheduleVariant = "WTODailyA"

	// ScheduleDailyDrivers is the variant of the variable daily working time
	// declaration for the employees who drive.
	ScheduleDailyDrivers ScheduleVariant = "WTODailyD"

	// ScheduleWeekly is the declaration of a fixed weekly working time.
	ScheduleWeekly ScheduleVariant = "WTOWeek"

	// ScheduleOvertime is the declaration of the overtime, expressed as a
	// working time.
	ScheduleOvertime ScheduleVariant = "WTOOv"

	// ScheduleOvertimeRetrospective is the retrospective variant of the
	// overtime working time declaration.
	ScheduleOvertimeRetrospective ScheduleVariant = "WTOOvA"

	// ScheduleOvertimeDrivers is the variant of the overtime working time
	// declaration for the employees who drive.
	ScheduleOvertimeDrivers ScheduleVariant = "WTOOvD"
)

// WorkStatusChangeDeclaration is the declaration of a change of the
// employment relationship for the working time organisation.
//
// It is the vehicle of the census of 2022, and it reports the working time
// of the employees of one business branch.
type WorkStatusChangeDeclaration struct {
	// BranchNumber is the number of the branch of the employer.
	BranchNumber string

	// RelatedProtocol refers to another declaration this one is related to.
	RelatedProtocol string

	// RelatedDate is the date of the declaration RelatedProtocol refers to.
	RelatedDate time.Time

	// Comments is an optional note for the reviewer of the declaration.
	Comments string

	// Employees are the changes of the employees of the branch.
	Employees []WorkStatusChange
}

// WorkStatusChange is the change of the working time of one employee.
type WorkStatusChange struct {
	// EmployeeTaxID is the tax identification number of the employee.
	EmployeeTaxID string

	// EmployeeFirstName is the first name of the employee.
	EmployeeFirstName string

	// EmployeeLastName is the last name of the employee.
	EmployeeLastName string

	// Date is the day the change applies from.
	Date time.Time

	// UsesDigitalWorkingTime reports whether the working time of the
	// employee is organised digitally.
	UsesDigitalWorkingTime bool

	// FullEmploymentHours are the weekly hours of the employment.
	FullEmploymentHours float64

	// WeeklyPattern is the number of working days of the week.
	WeeklyPattern WeekPattern

	// FlexibleArrivalMinutes are the minutes the employee may arrive
	// earlier.
	FlexibleArrivalMinutes int

	// UsesWorkCard reports whether the employee stamps the work card.
	UsesWorkCard bool

	// BreakMinutes are the minutes of the break of the employee.
	BreakMinutes int

	// BreakWithinSchedule reports whether the break is inside the schedule.
	BreakWithinSchedule bool
}

// ExemptionDeclaration is the declaration of an exemption from the duty of
// advance notification.
//
// The branches of the digital work card regime file it every month.
type ExemptionDeclaration struct {
	// BranchNumber is the number of the branch of the employer.
	BranchNumber string

	// Excluded reports whether the branch is exempt.
	Excluded bool

	// Month is the month the exemption covers.
	Month int

	// Year is the year the exemption covers.
	Year int

	// Comments is an optional note for the reviewer of the declaration.
	Comments string
}

// SubmitWorkStatusChange submits the declaration of a change of the
// employment relationship for the working time organisation.
func (c *Client) SubmitWorkStatusChange(ctx context.Context, declaration WorkStatusChangeDeclaration) ([]Submission, error) {
	payload := &wire.WorkingStatusChangePayload{}

	branch := wire.CompanyWorkingStatusChange{
		BranchNumber:    declaration.BranchNumber,
		RelatedProtocol: declaration.RelatedProtocol,
		RelatedDate:     wire.Date{Time: declaration.RelatedDate},
		Comments:        declaration.Comments,
	}

	for i, employee := range declaration.Employees {
		field := fmt.Sprintf("declarations.employees[%d]", i)

		if err := requireAFM(field+".employeeTaxID", employee.EmployeeTaxID); err != nil {
			return nil, err
		}
		if employee.Date.IsZero() {
			return nil, &ValidationError{Field: field + ".date", Reason: "the date of the change is required"}
		}

		branch.Employees.Employees = append(branch.Employees.Employees, wire.WorkingStatusChangeEmployee{
			TaxID:                  employee.EmployeeTaxID,
			LastName:               employee.EmployeeLastName,
			FirstName:              employee.EmployeeFirstName,
			Date:                   wire.Date{Time: employee.Date},
			UsesDigitalWorkingTime: wire.FromBool(employee.UsesDigitalWorkingTime),
			FullEmploymentHours:    wire.Hours(employee.FullEmploymentHours),
			WeeklyPattern:          weeklyPatternCode(employee.WeeklyPattern),
			FlexibleArrivalMinutes: wire.Minutes(employee.FlexibleArrivalMinutes),
			UsesWorkCard:           wire.FromBool(employee.UsesWorkCard),
			BreakMinutes:           wire.Minutes(employee.BreakMinutes),
			BreakWithinSchedule:    wire.FromBool(employee.BreakWithinSchedule),
		})
	}

	payload.WorkingStatusChanges.WorkingStatusChange = append(
		payload.WorkingStatusChanges.WorkingStatusChange, branch)

	return c.submit(ctx, CodeWorkStatusChange, payload)
}

// SubmitExemptionNotification submits the declaration of an exemption from
// the duty of advance notification.
func (c *Client) SubmitExemptionNotification(ctx context.Context, declaration ExemptionDeclaration) ([]Submission, error) {
	if err := requireBranch("declarations.branchNumber", declaration.BranchNumber); err != nil {
		return nil, err
	}
	if declaration.Month < 1 || declaration.Month > 12 {
		return nil, &ValidationError{Field: "declarations.month", Reason: "the month must be between 1 and 12"}
	}
	if declaration.Year < 1900 || declaration.Year > 2200 {
		return nil, &ValidationError{Field: "declarations.year", Reason: "the year is not a plausible one"}
	}

	payload := &wire.ExemptionPayload{}
	payload.Exemptions.Exemption = append(payload.Exemptions.Exemption, wire.CompanyExemption{
		BranchNumber: declaration.BranchNumber,
		Excluded:     wire.FromBool(declaration.Excluded),
		Month:        declaration.Month,
		Year:         declaration.Year,
		Comments:     declaration.Comments,
	})

	return c.submit(ctx, CodeExemptionNotification, payload)
}

// weeklyPatternCode returns the code the API uses for the weekly pattern.
func weeklyPatternCode(pattern WeekPattern) string {
	switch pattern {
	case WeekFiveDays:
		return "5"
	case WeekSixDays:
		return "6"
	default:
		return ""
	}
}
