package ergani

import "encoding/json"

// The Ergani API expects several payloads to be nested one level deeper than a
// flat struct can express. A field documented as "Details>CardDetails" has to
// be sent as:
//
//	{"Details": {"CardDetails": [...]}}
//
// The standard library's encoding/json has no syntax for that: a tag such as
// json:"Details>CardDetails" is emitted verbatim as a single literal key
// ("Details>CardDetails"), which the API rejects.
//
// The container types therefore implement json.Marshaler themselves. The
// payload structs below mirror the documented field names and their order, and
// the MarshalJSON methods assemble the nested objects.

// workCardDetailsPayload wraps the work card entries of a single business branch.
type workCardDetailsPayload struct {
	CardDetails []WorkCard `json:"CardDetails"`
}

// companyWorkCardPayload is the wire representation of a CompanyWorkCard.
type companyWorkCardPayload struct {
	EmployerTaxID        string                 `json:"f_afm_ergodoti"`
	BusinessBranchNumber int                    `json:"f_aa"`
	Comments             string                 `json:"f_comments,omitempty"`
	Details              workCardDetailsPayload `json:"Details"`
}

// MarshalJSON encodes a CompanyWorkCard with its work cards nested under
// "Details" as the API expects.
func (c CompanyWorkCard) MarshalJSON() ([]byte, error) {
	details := c.CardDetails
	if details == nil {
		details = []WorkCard{}
	}

	return json.Marshal(companyWorkCardPayload{
		EmployerTaxID:        c.EmployerTaxID,
		BusinessBranchNumber: c.BusinessBranchNumber,
		Comments:             c.Comments,
		Details:              workCardDetailsPayload{CardDetails: details},
	})
}

// overtimeEmployeesPayload wraps the overtime entries of a single business branch.
type overtimeEmployeesPayload struct {
	OvertimeErgazomenosDate []Overtime `json:"OvertimeErgazomenosDate"`
}

// companyOvertimePayload is the wire representation of a CompanyOvertime.
type companyOvertimePayload struct {
	BusinessBranchNumber   int                      `json:"f_aa_pararthmatos"`
	SEPEServiceCode        string                   `json:"f_ypiresia_sepe"`
	PrimaryActivityCode    string                   `json:"f_kad_kyria"`
	BranchActivityCode     string                   `json:"f_kad_pararthmatos"`
	KallikratisCode        string                   `json:"f_kallikratis_pararthmatos"`
	LegalRepTaxID          string                   `json:"f_afm_proswpoy"`
	EmployeeOvertimes      overtimeEmployeesPayload `json:"Ergazomenoi"`
	RelatedProtocolID      string                   `json:"f_rel_protocol,omitempty"`
	RelatedProtocolDate    *Date                    `json:"f_rel_date,omitempty"`
	EmployerOrganization   string                   `json:"f_ergodotikh_organwsh,omitempty"`
	SecondaryActivityCode1 string                   `json:"f_kad_deyt_1,omitempty"`
	SecondaryActivityCode2 string                   `json:"f_kad_deyt_2,omitempty"`
	SecondaryActivityCode3 string                   `json:"f_kad_deyt_3,omitempty"`
	SecondaryActivityCode4 string                   `json:"f_kad_deyt_4,omitempty"`
	Comments               string                   `json:"f_comments,omitempty"`
}

// MarshalJSON encodes a CompanyOvertime with its overtime entries nested under
// "Ergazomenoi" as the API expects.
func (c CompanyOvertime) MarshalJSON() ([]byte, error) {
	overtimes := c.EmployeeOvertimes
	if overtimes == nil {
		overtimes = []Overtime{}
	}

	return json.Marshal(companyOvertimePayload{
		BusinessBranchNumber:   c.BusinessBranchNumber,
		SEPEServiceCode:        c.SEPEServiceCode,
		PrimaryActivityCode:    c.PrimaryActivityCode,
		BranchActivityCode:     c.BranchActivityCode,
		KallikratisCode:        c.KallikratisCode,
		LegalRepTaxID:          c.LegalRepTaxID,
		EmployeeOvertimes:      overtimeEmployeesPayload{OvertimeErgazomenosDate: overtimes},
		RelatedProtocolID:      c.RelatedProtocolID,
		RelatedProtocolDate:    c.RelatedProtocolDate,
		EmployerOrganization:   c.EmployerOrganization,
		SecondaryActivityCode1: c.SecondaryActivityCode1,
		SecondaryActivityCode2: c.SecondaryActivityCode2,
		SecondaryActivityCode3: c.SecondaryActivityCode3,
		SecondaryActivityCode4: c.SecondaryActivityCode4,
		Comments:               c.Comments,
	})
}

// workdayAnalyticsPayload wraps the working hours of a single employee schedule.
type workdayAnalyticsPayload struct {
	ErgazomenosWTOAnalytics []WorkdayDetails `json:"ErgazomenosWTOAnalytics"`
}

// employeeDailySchedulePayload is the wire representation of an EmployeeDailySchedule.
type employeeDailySchedulePayload struct {
	EmployeeTaxID     string                  `json:"f_afm"`
	EmployeeLastName  string                  `json:"f_eponymo"`
	EmployeeFirstName string                  `json:"f_onoma"`
	ScheduleDate      Date                    `json:"f_date"`
	Analytics         workdayAnalyticsPayload `json:"ErgazomenosAnalytics"`
}

// MarshalJSON encodes an EmployeeDailySchedule with its working hours nested
// under "ErgazomenosAnalytics" as the API expects.
func (e EmployeeDailySchedule) MarshalJSON() ([]byte, error) {
	details := e.WorkdayDetails
	if details == nil {
		details = []WorkdayDetails{}
	}

	return json.Marshal(employeeDailySchedulePayload{
		EmployeeTaxID:     e.EmployeeTaxID,
		EmployeeLastName:  e.EmployeeLastName,
		EmployeeFirstName: e.EmployeeFirstName,
		ScheduleDate:      e.ScheduleDate,
		Analytics:         workdayAnalyticsPayload{ErgazomenosWTOAnalytics: details},
	})
}

// employeeWeeklySchedulePayload is the wire representation of an EmployeeWeeklySchedule.
type employeeWeeklySchedulePayload struct {
	EmployeeTaxID     string                  `json:"f_afm"`
	EmployeeLastName  string                  `json:"f_eponymo"`
	EmployeeFirstName string                  `json:"f_onoma"`
	ScheduleDay       Weekday                 `json:"f_day"`
	Analytics         workdayAnalyticsPayload `json:"ErgazomenosAnalytics"`
}

// MarshalJSON encodes an EmployeeWeeklySchedule with its working hours nested
// under "ErgazomenosAnalytics" as the API expects.
func (e EmployeeWeeklySchedule) MarshalJSON() ([]byte, error) {
	details := e.WorkdayDetails
	if details == nil {
		details = []WorkdayDetails{}
	}

	return json.Marshal(employeeWeeklySchedulePayload{
		EmployeeTaxID:     e.EmployeeTaxID,
		EmployeeLastName:  e.EmployeeLastName,
		EmployeeFirstName: e.EmployeeFirstName,
		ScheduleDay:       e.ScheduleDay,
		Analytics:         workdayAnalyticsPayload{ErgazomenosWTOAnalytics: details},
	})
}

// dailyScheduleEmployeesPayload wraps the daily employee schedules of a business branch.
type dailyScheduleEmployeesPayload struct {
	ErgazomenoiWTO []EmployeeDailySchedule `json:"ErgazomenoiWTO"`
}

// companyDailySchedulePayload is the wire representation of a CompanyDailySchedule.
type companyDailySchedulePayload struct {
	BusinessBranchNumber int                           `json:"f_aa_pararthmatos"`
	StartDate            *Date                         `json:"f_from_date,omitempty"`
	EndDate              *Date                         `json:"f_to_date,omitempty"`
	EmployeeSchedules    dailyScheduleEmployeesPayload `json:"Ergazomenoi"`
	RelatedProtocolID    string                        `json:"f_rel_protocol,omitempty"`
	RelatedProtocolDate  *Date                         `json:"f_rel_date,omitempty"`
	Comments             string                        `json:"f_comments,omitempty"`
}

// MarshalJSON encodes a CompanyDailySchedule with its employee schedules nested
// under "Ergazomenoi" as the API expects.
func (c CompanyDailySchedule) MarshalJSON() ([]byte, error) {
	schedules := c.EmployeeSchedules
	if schedules == nil {
		schedules = []EmployeeDailySchedule{}
	}

	return json.Marshal(companyDailySchedulePayload{
		BusinessBranchNumber: c.BusinessBranchNumber,
		StartDate:            c.StartDate,
		EndDate:              c.EndDate,
		EmployeeSchedules:    dailyScheduleEmployeesPayload{ErgazomenoiWTO: schedules},
		RelatedProtocolID:    c.RelatedProtocolID,
		RelatedProtocolDate:  c.RelatedProtocolDate,
		Comments:             c.Comments,
	})
}

// weeklyScheduleEmployeesPayload wraps the weekly employee schedules of a business branch.
type weeklyScheduleEmployeesPayload struct {
	ErgazomenoiWTO []EmployeeWeeklySchedule `json:"ErgazomenoiWTO"`
}

// companyWeeklySchedulePayload is the wire representation of a CompanyWeeklySchedule.
type companyWeeklySchedulePayload struct {
	BusinessBranchNumber int                            `json:"f_aa_pararthmatos"`
	StartDate            Date                           `json:"f_from_date"`
	EndDate              Date                           `json:"f_to_date"`
	EmployeeSchedules    weeklyScheduleEmployeesPayload `json:"Ergazomenoi"`
	RelatedProtocolID    string                         `json:"f_rel_protocol,omitempty"`
	RelatedProtocolDate  *Date                          `json:"f_rel_date,omitempty"`
	Comments             string                         `json:"f_comments,omitempty"`
}

// MarshalJSON encodes a CompanyWeeklySchedule with its employee schedules nested
// under "Ergazomenoi" as the API expects.
func (c CompanyWeeklySchedule) MarshalJSON() ([]byte, error) {
	schedules := c.EmployeeSchedules
	if schedules == nil {
		schedules = []EmployeeWeeklySchedule{}
	}

	return json.Marshal(companyWeeklySchedulePayload{
		BusinessBranchNumber: c.BusinessBranchNumber,
		StartDate:            c.StartDate,
		EndDate:              c.EndDate,
		EmployeeSchedules:    weeklyScheduleEmployeesPayload{ErgazomenoiWTO: schedules},
		RelatedProtocolID:    c.RelatedProtocolID,
		RelatedProtocolDate:  c.RelatedProtocolDate,
		Comments:             c.Comments,
	})
}
