package wire

// ScheduleDeclarations is the root of a working time organisation
// submission.
//
// The same root serves the daily schedule (WTODaily), the weekly schedule
// (WTOWeek), the overtime schedule (WTOOv) and the leaves (wtoHoliday). The
// schemas are WTO_v2.xsd and WTO_Weekly_v1.xsd.
type ScheduleDeclarations struct {
	WTOS ScheduleList `json:"WTOS"`
}

// ScheduleList holds the declarations of the submission.
type ScheduleList struct {
	WTO []CompanySchedule `json:"WTO"`
}

// CompanySchedule groups the schedules of one business branch.
type CompanySchedule struct {
	BranchNumber    string               `json:"f_aa_pararthmatos"`
	RelatedProtocol string               `json:"f_rel_protocol"`
	RelatedDate     Date                 `json:"f_rel_date"`
	Comments        string               `json:"f_comments"`
	FromDate        Date                 `json:"f_from_date"`
	ToDate          Date                 `json:"f_to_date"`
	Employees       EmployeeScheduleList `json:"Ergazomenoi"`
}

// EmployeeScheduleList holds the schedules of the employees of a branch.
type EmployeeScheduleList struct {
	WTO []EmployeeSchedule `json:"ErgazomenoiWTO"`
}

// EmployeeSchedule is the schedule of one employee.
//
// The daily, the weekly and the leave declarations differ only in the field
// that identifies the day: f_date for the daily one and for the leaves,
// f_day for the weekly one.
type EmployeeSchedule struct {
	EmployeeTaxID     string           `json:"f_afm"`
	EmployeeLastName  string           `json:"f_eponymo"`
	EmployeeFirstName string           `json:"f_onoma"`
	Date              *Date            `json:"f_date,omitempty"`
	Day               *Day             `json:"f_day,omitempty"`
	Analytics         WorkdayAnalytics `json:"ErgazomenosAnalytics"`
}

// WorkdayAnalytics holds the working periods of one employee.
type WorkdayAnalytics struct {
	Analyses []WorkdayAnalysis `json:"ErgazomenosWTOAnalytics"`
}

// WorkdayAnalysis is one working period.
//
// The year and the entitled leave days only apply to the leave declarations.
type WorkdayAnalysis struct {
	WorkType     string `json:"f_type"`
	FromTime     Time   `json:"f_from"`
	ToTime       Time   `json:"f_to"`
	Year         string `json:"f_year,omitempty"`
	EntitledDays string `json:"f_req_days,omitempty"`
}
