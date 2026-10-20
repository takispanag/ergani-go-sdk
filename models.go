package ergani

import (
	"fmt"
	"time"
)

// The models of this package carry no JSON tags and no knowledge of the wire
// format. They are plain data: the zero value of every field means "not set",
// and the SDK turns them into the payload the API expects.

// WorkCardDeclaration groups the work cards of one business branch of one
// employer.
//
// It is submitted with [Client.SubmitWorkCards].
type WorkCardDeclaration struct {
	// EmployerTaxID is the tax identification number of the employer. The
	// API checks it against the connected employer.
	EmployerTaxID string

	// BranchNumber is the number of the branch of the employer. The main
	// branch is "0".
	BranchNumber string

	// Comments is an optional note for the reviewer of the declaration.
	Comments string

	// Cards are the movements of the employees of the branch.
	Cards []WorkCard
}

// WorkCard is one movement of one employee.
type WorkCard struct {
	// EmployeeTaxID is the tax identification number of the employee.
	EmployeeTaxID string

	// EmployeeFirstName is the first name of the employee.
	EmployeeFirstName string

	// EmployeeLastName is the last name of the employee.
	EmployeeLastName string

	// MovementType is the kind of movement.
	MovementType MovementType

	// ReferenceDate is the day the declaration refers to. It is sent in the
	// YYYY-MM-DD format.
	ReferenceDate time.Time

	// MovementDateTime is the moment of the movement.
	MovementDateTime time.Time

	// LateJustification explains why the movement is declared late. It is
	// nil when the movement is declared on time.
	LateJustification *LateJustification
}

// OvertimeDeclaration groups the overtime of one business branch of one
// employer.
//
// It is submitted with [Client.SubmitOvertime].
type OvertimeDeclaration struct {
	// BranchNumber is the number of the branch of the employer.
	BranchNumber string

	// SEPEServiceCode is the code of the labour inspection service of the
	// branch.
	SEPEServiceCode string

	// PrimaryActivityCode is the primary activity code of the branch.
	PrimaryActivityCode string

	// SecondaryActivityCode1 to SecondaryActivityCode4 are the secondary
	// activity codes of the branch.
	SecondaryActivityCode1 string
	SecondaryActivityCode2 string
	SecondaryActivityCode3 string
	SecondaryActivityCode4 string

	// BranchActivityCode is the activity code of the branch.
	BranchActivityCode string

	// MunicipalCode is the municipal code of the branch.
	MunicipalCode string

	// LegalRepTaxID is the tax identification number of the legal
	// representative of the employer.
	LegalRepTaxID string

	// EmployerOrganization is an optional note about the employer.
	EmployerOrganization string

	// RelatedProtocol refers to another declaration this one is related to.
	RelatedProtocol string

	// RelatedDate is the date of the declaration RelatedProtocol refers to.
	RelatedDate time.Time

	// Comments is an optional note for the reviewer of the declaration.
	Comments string

	// Records are the overtime of the employees of the branch.
	Records []OvertimeRecord
}

// OvertimeRecord is the overtime of one employee on one day.
type OvertimeRecord struct {
	// EmployeeTaxID is the tax identification number of the employee.
	EmployeeTaxID string

	// EmployeeSSN is the social security registry number of the employee.
	EmployeeSSN string

	// EmployeeFirstName is the first name of the employee.
	EmployeeFirstName string

	// EmployeeLastName is the last name of the employee.
	EmployeeLastName string

	// Date is the day the overtime happened on.
	Date time.Time

	// StartTime is the moment the overtime started. Only the hour and the
	// minute are sent to the API.
	StartTime time.Time

	// EndTime is the moment the overtime ended. Only the hour and the minute
	// are sent to the API.
	EndTime time.Time

	// SecondStartTime and SecondEndTime are the second interval of a split
	// overtime. They are left out when they are not set.
	SecondStartTime time.Time
	SecondEndTime   time.Time

	// Cancelled reports whether the overtime was cancelled.
	Cancelled bool

	// ProfessionCode is the code of the profession of the employee.
	ProfessionCode string

	// Justification is the official reason for the overtime.
	Justification OvertimeJustification

	// WeekPattern is the number of working days of the week of the
	// employee.
	WeekPattern WeekPattern

	// ASEEApproval is the reference of the approval of the relevant
	// authority, when the overtime needs one.
	ASEEApproval string
}

// DailyScheduleDeclaration groups the daily schedules of one business branch
// of one employer.
//
// It is submitted with [Client.SubmitDailySchedules].
type DailyScheduleDeclaration struct {
	// BranchNumber is the number of the branch of the employer.
	BranchNumber string

	// StartDate and EndDate bound the period the declaration covers. They
	// are left out when they are not set.
	StartDate time.Time
	EndDate   time.Time

	// RelatedProtocol refers to another declaration this one is related to.
	RelatedProtocol string

	// RelatedDate is the date of the declaration RelatedProtocol refers to.
	RelatedDate time.Time

	// Comments is an optional note for the reviewer of the declaration.
	Comments string

	// Schedules are the schedules of the employees of the branch.
	Schedules []EmployeeDailySchedule
}

// EmployeeDailySchedule is the working schedule of one employee for one day.
type EmployeeDailySchedule struct {
	// EmployeeTaxID is the tax identification number of the employee.
	EmployeeTaxID string

	// EmployeeFirstName is the first name of the employee.
	EmployeeFirstName string

	// EmployeeLastName is the last name of the employee.
	EmployeeLastName string

	// Date is the day the schedule applies to.
	Date time.Time

	// Workday are the working periods of the day. A day may hold more than
	// one, for a split shift.
	Workday []Workday
}

// WeeklyScheduleDeclaration groups the weekly schedules of one business
// branch of one employer.
//
// It is submitted with [Client.SubmitWeeklySchedules].
type WeeklyScheduleDeclaration struct {
	// BranchNumber is the number of the branch of the employer.
	BranchNumber string

	// StartDate is the first day of the week.
	StartDate time.Time

	// EndDate is the last day of the week.
	EndDate time.Time

	// RelatedProtocol refers to another declaration this one is related to.
	RelatedProtocol string

	// RelatedDate is the date of the declaration RelatedProtocol refers to.
	RelatedDate time.Time

	// Comments is an optional note for the reviewer of the declaration.
	Comments string

	// Schedules are the schedules of the employees of the branch.
	Schedules []EmployeeWeeklySchedule
}

// EmployeeWeeklySchedule is the working schedule of one employee for one day
// of a week.
type EmployeeWeeklySchedule struct {
	// EmployeeTaxID is the tax identification number of the employee.
	EmployeeTaxID string

	// EmployeeFirstName is the first name of the employee.
	EmployeeFirstName string

	// EmployeeLastName is the last name of the employee.
	EmployeeLastName string

	// Day is the day of the week the schedule applies to. The API numbers
	// the days from Sunday, as time.Weekday does.
	Day time.Weekday

	// Workday are the working periods of the day.
	Workday []Workday
}

// OvertimeScheduleDeclaration groups the overtime working time declarations
// of one business branch of one employer.
//
// It is submitted with [Client.SubmitOvertimeSchedules].
type OvertimeScheduleDeclaration struct {
	// BranchNumber is the number of the branch of the employer.
	BranchNumber string

	// StartDate and EndDate bound the period the declaration covers.
	StartDate time.Time
	EndDate   time.Time

	// Comments is an optional note for the reviewer of the declaration.
	Comments string

	// Schedules are the schedules of the employees of the branch.
	Schedules []EmployeeOvertimeSchedule
}

// EmployeeOvertimeSchedule is the overtime of one employee on one day,
// expressed as a working time declaration.
type EmployeeOvertimeSchedule struct {
	// EmployeeTaxID is the tax identification number of the employee.
	EmployeeTaxID string

	// EmployeeFirstName is the first name of the employee.
	EmployeeFirstName string

	// EmployeeLastName is the last name of the employee.
	EmployeeLastName string

	// Date is the day the overtime happened on.
	Date time.Time

	// Workday are the overtime periods of the day.
	Workday []Workday
}

// LeaveDeclaration groups the leave declarations of one business branch of
// one employer.
//
// It is submitted with [Client.SubmitLeaves].
type LeaveDeclaration struct {
	// BranchNumber is the number of the branch of the employer.
	BranchNumber string

	// StartDate and EndDate bound the period the declaration covers.
	StartDate time.Time
	EndDate   time.Time

	// Comments is an optional note for the reviewer of the declaration.
	Comments string

	// Schedules are the leaves of the employees of the branch.
	Schedules []EmployeeLeave
}

// EmployeeLeave is the leave of one employee on one day.
type EmployeeLeave struct {
	// EmployeeTaxID is the tax identification number of the employee.
	EmployeeTaxID string

	// EmployeeFirstName is the first name of the employee.
	EmployeeFirstName string

	// EmployeeLastName is the last name of the employee.
	EmployeeLastName string

	// Date is the day the leave applies to.
	Date time.Time

	// LeaveType is the kind of leave.
	LeaveType LeaveType

	// Year is the year the leave belongs to, for the entitlements that are
	// counted per year. It is left out when it is not set.
	Year int

	// EntitledDays is the number of days the employee is entitled to for
	// this kind of leave. It is left out when it is not set.
	EntitledDays int
}

// Workday is a working period within a day.
//
// The same type serves the daily and the weekly schedules, the overtime
// schedules and the leaves.
type Workday struct {
	// WorkType is the kind of work.
	WorkType WorkType

	// StartTime is the moment the work starts. Only the hour and the minute
	// are sent to the API.
	StartTime time.Time

	// EndTime is the moment the work ends. Only the hour and the minute are
	// sent to the API.
	EndTime time.Time
}

// Submission is the acknowledgement of an accepted declaration.
type Submission struct {
	// ID is the identifier the API gave to the declaration.
	ID string

	// Protocol is the protocol number of the declaration. It is what the
	// API expects to refer to the declaration later.
	Protocol string

	// SubmittedAt is the moment the API recorded the declaration.
	SubmittedAt time.Time
}

// SubmissionType is a type of declaration the connected employer may submit.
type SubmissionType struct {
	// ID is the identifier of the declaration type.
	ID int

	// Code is the code of the declaration type. It is the last segment of
	// the path of a submission.
	Code DocumentCode

	// Description is the Greek description of the declaration type.
	Description string
}

// DocumentID returns the identifier of the declaration type in the form the
// API expects it in the recall of a declaration: the number, zero padded to
// five digits, such as "00009".
func (t SubmissionType) DocumentID() string {
	return fmt.Sprintf("%05d", t.ID)
}
