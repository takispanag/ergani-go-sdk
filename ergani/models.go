package ergani

import (
	"encoding/json"
	"time"
)

// WorkCard represents a single work card event (arrival or departure) for an employee.
type WorkCard struct {
	EmployeeTaxID                string                            `json:"f_afm"`
	EmployeeLastName             string                            `json:"f_eponymo"`
	EmployeeFirstName            string                            `json:"f_onoma"`
	WorkCardMovementType         WorkCardMovementType              `json:"f_type"`
	WorkCardSubmissionDate       Date                              `json:"f_reference_date"`
	WorkCardMovementDateTime     DateTime                          `json:"f_date"`
	LateDeclarationJustification *LateDeclarationJustificationType `json:"f_aitiologia,omitempty"`
}

// CompanyWorkCard groups work card entries for a single business branch.
type CompanyWorkCard struct {
	EmployerTaxID        string `json:"f_afm_ergodoti"`
	BusinessBranchNumber int    `json:"f_aa"`
	Comments             string `json:"f_comments,omitempty"`
	// CardDetails are nested within "Details>CardDetails" in the final JSON.
	CardDetails []WorkCard `json:"Details>CardDetails"`
}

// Overtime represents an overtime entry for an employee on a specific date.
type Overtime struct {
	EmployeeTaxID          string                    `json:"f_afm"`
	EmployeeSSN            string                    `json:"f_amka"`
	EmployeeLastName       string                    `json:"f_eponymo"`
	EmployeeFirstName      string                    `json:"f_onoma"`
	OvertimeDate           Date                      `json:"f_date"`
	OvertimeStartTime      Time                      `json:"f_from"`
	OvertimeEndTime        Time                      `json:"f_to"`
	OvertimeCancellation   Bool                      `json:"f_cancellation"`
	EmployeeProfessionCode string                    `json:"f_step"`
	OvertimeJustification  OvertimeJustificationType `json:"f_reason"`
	WeeklyWorkdaysNumber   int                       `json:"f_weekdates"` // Valid values are 5 or 6
	ASEEApproval           string                    `json:"f_asee,omitempty"`
}

// CompanyOvertime groups overtime entries for a single business branch.
type CompanyOvertime struct {
	BusinessBranchNumber int    `json:"f_aa_pararthmatos"`
	SEPEServiceCode      string `json:"f_ypiresia_sepe"`
	PrimaryActivityCode  string `json:"f_kad_kyria"`
	BranchActivityCode   string `json:"f_kad_pararthmatos"`
	KallikratisCode      string `json:"f_kallikratis_pararthmatos"`
	LegalRepTaxID        string `json:"f_afm_proswpoy"`
	// EmployeeOvertimes are nested within "Ergazomenoi>OvertimeErgazomenosDate".
	EmployeeOvertimes      []Overtime `json:"Ergazomenoi>OvertimeErgazomenosDate"`
	RelatedProtocolID      string     `json:"f_rel_protocol,omitempty"`
	RelatedProtocolDate    *Date      `json:"f_rel_date,omitempty"`
	EmployerOrganization   string     `json:"f_ergodotikh_organwsh,omitempty"`
	SecondaryActivityCode1 string     `json:"f_kad_deyt_1,omitempty"`
	SecondaryActivityCode2 string     `json:"f_kad_deyt_2,omitempty"`
	SecondaryActivityCode3 string     `json:"f_kad_deyt_3,omitempty"`
	SecondaryActivityCode4 string     `json:"f_kad_deyt_4,omitempty"`
	Comments               string     `json:"f_comments,omitempty"`
}

// WorkdayDetails represents the working hours and type of work for a part of a day.
// A day can have multiple such entries (e.g., split shift).
type WorkdayDetails struct {
	WorkType  ScheduleWorkType `json:"f_type"`
	StartTime Time             `json:"f_from"`
	EndTime   Time             `json:"f_to"`
}

// EmployeeDailySchedule represents the complete work schedule for an employee for one day.
type EmployeeDailySchedule struct {
	EmployeeTaxID     string `json:"f_afm"`
	EmployeeLastName  string `json:"f_eponymo"`
	EmployeeFirstName string `json:"f_onoma"`
	ScheduleDate      Date   `json:"f_date"`
	// WorkdayDetails are nested within "ErgazomenosAnalytics>ErgazomenosWTOAnalytics".
	WorkdayDetails []WorkdayDetails `json:"ErgazomenosAnalytics>ErgazomenosWTOAnalytics"`
}

// CompanyDailySchedule groups daily employee schedules for a business branch.
type CompanyDailySchedule struct {
	BusinessBranchNumber int   `json:"f_aa_pararthmatos"`
	StartDate            *Date `json:"f_from_date,omitempty"`
	EndDate              *Date `json:"f_to_date,omitempty"`
	// EmployeeSchedules are nested within "Ergazomenoi>ErgazomenoiWTO".
	EmployeeSchedules   []EmployeeDailySchedule `json:"Ergazomenoi>ErgazomenoiWTO"`
	RelatedProtocolID   string                  `json:"f_rel_protocol,omitempty"`
	RelatedProtocolDate *Date                   `json:"f_rel_date,omitempty"`
	Comments            string                  `json:"f_comments,omitempty"`
}

// EmployeeWeeklySchedule represents a weekly schedule entry for an employee for a specific weekday.
type EmployeeWeeklySchedule struct {
	EmployeeTaxID     string  `json:"f_afm"`
	EmployeeLastName  string  `json:"f_eponymo"`
	EmployeeFirstName string  `json:"f_onoma"`
	ScheduleDay       Weekday `json:"f_day"`
	// WorkdayDetails are nested within "ErgazomenosAnalytics>ErgazomenosWTOAnalytics".
	WorkdayDetails []WorkdayDetails `json:"ErgazomenosAnalytics>ErgazomenosWTOAnalytics"`
}

// CompanyWeeklySchedule groups weekly employee schedules for a business branch
// over a specified date range.
type CompanyWeeklySchedule struct {
	BusinessBranchNumber int  `json:"f_aa_pararthmatos"`
	StartDate            Date `json:"f_from_date"`
	EndDate              Date `json:"f_to_date"`
	// EmployeeSchedules are nested within "Ergazomenoi>ErgazomenoiWTO".
	EmployeeSchedules   []EmployeeWeeklySchedule `json:"Ergazomenoi>ErgazomenoiWTO"`
	RelatedProtocolID   string                   `json:"f_rel_protocol,omitempty"`
	RelatedProtocolDate *Date                    `json:"f_rel_date,omitempty"`
	Comments            string                   `json:"f_comments,omitempty"`
}

// SubmissionResponse represents the data returned from a successful submission to the API.
type SubmissionResponse struct {
	ID       string `json:"id"`
	Protocol string `json:"protocol"`
	// SubmissionDate is parsed from a custom format and not directly from JSON.
	SubmissionDate time.Time `json:"-"`
}

// UnmarshalJSON handles the custom date format ("02/01/2006 15:04") from the API
// for the SubmissionDate field.
func (s *SubmissionResponse) UnmarshalJSON(data []byte) error {
	type Alias SubmissionResponse
	aux := &struct {
		ID                string `json:"id"`
		Protocol          string `json:"protocol"`
		SubmissionDateStr string `json:"submitDate"`
		*Alias
	}{
		Alias: (*Alias)(s),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Set all fields properly
	s.ID = aux.ID
	s.Protocol = aux.Protocol

	// Parse the custom date string
	t, err := time.Parse("02/01/2006 15:04", aux.SubmissionDateStr)
	if err != nil {
		return err
	}

	s.SubmissionDate = t
	return nil
}
