package wire

// WorkingStatusChangePayload is the root of a declaration of a change of the
// employment relationship for the working time organisation.
//
// The schema is WorkingStatusChange_v1.xsd.
type WorkingStatusChangePayload struct {
	WorkingStatusChanges WorkingStatusChangeList `json:"WorkingStatusChanges"`
}

// WorkingStatusChangeList holds the declarations of the submission.
type WorkingStatusChangeList struct {
	WorkingStatusChange []CompanyWorkingStatusChange `json:"WorkingStatusChange"`
}

// CompanyWorkingStatusChange groups the changes of one business branch.
type CompanyWorkingStatusChange struct {
	BranchNumber    string                          `json:"f_aa_pararthmatos"`
	RelatedProtocol string                          `json:"f_rel_protocol"`
	RelatedDate     Date                            `json:"f_rel_date"`
	Comments        string                          `json:"f_comments"`
	Employees       WorkingStatusChangeEmployeeList `json:"Ergazomenoi"`
}

// WorkingStatusChangeEmployeeList holds the employees of a branch.
type WorkingStatusChangeEmployeeList struct {
	Employees []WorkingStatusChangeEmployee `json:"Ergazomenos"`
}

// WorkingStatusChangeEmployee is the change of the working time of one
// employee.
type WorkingStatusChangeEmployee struct {
	TaxID                  string  `json:"f_afm"`
	LastName               string  `json:"f_eponymo"`
	FirstName              string  `json:"f_onoma"`
	Date                   Date    `json:"f_date"`
	UsesDigitalWorkingTime Bool    `json:"f_working_time_digital_organization"`
	FullEmploymentHours    Hours   `json:"f_full_employment_hours"`
	WeeklyPattern          string  `json:"f_week_days"`
	FlexibleArrivalMinutes Minutes `json:"f_euelikto_wrario_minutes"`
	UsesWorkCard           Bool    `json:"f_working_card"`
	BreakMinutes           Minutes `json:"f_dialeimma_minutes"`
	BreakWithinSchedule    Bool    `json:"f_dialeimma_entos_wrariou"`
}

// ExemptionPayload is the root of a declaration of an exemption from the duty
// of advance notification.
//
// The schema is ExProan.xsd.
type ExemptionPayload struct {
	Exemptions ExemptionList `json:"DhlwseisExProan"`
}

// ExemptionList holds the declarations of the submission.
type ExemptionList struct {
	Exemption []CompanyExemption `json:"DhlwshExProan"`
}

// CompanyExemption is the exemption of one business branch for one month.
type CompanyExemption struct {
	BranchNumber string `json:"f_aa_pararthmatos"`
	Excluded     Bool   `json:"f_is_excluded"`
	Month        int    `json:"f_month"`
	Year         int    `json:"f_year"`
	Comments     string `json:"f_comments"`
}
