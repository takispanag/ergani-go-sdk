package wire

// OvertimeDeclarations is the root of an E8 overtime submission.
//
// The schema is EOvertime_v2.xsd. The root holds the overtime of every
// business branch of the employer.
type OvertimeDeclarations struct {
	Overtimes OvertimeList `json:"Overtimes"`
}

// OvertimeList holds the overtime declarations of the submission.
type OvertimeList struct {
	Overtime []CompanyOvertime `json:"Overtime"`
}

// CompanyOvertime groups the overtime of one business branch.
//
// The branch carries the codes of the employer that the API validates the
// submission against.
type CompanyOvertime struct {
	BranchNumber           string               `json:"f_aa_pararthmatos"`
	RelatedProtocol        string               `json:"f_rel_protocol"`
	RelatedDate            Date                 `json:"f_rel_date"`
	SEPEServiceCode        string               `json:"f_ypiresia_sepe"`
	EmployerOrganization   string               `json:"f_ergodotikh_organwsh"`
	PrimaryActivityCode    string               `json:"f_kad_kyria"`
	SecondaryActivityCode1 string               `json:"f_kad_deyt_1"`
	SecondaryActivityCode2 string               `json:"f_kad_deyt_2"`
	SecondaryActivityCode3 string               `json:"f_kad_deyt_3"`
	SecondaryActivityCode4 string               `json:"f_kad_deyt_4"`
	BranchActivityCode     string               `json:"f_kad_pararthmatos"`
	MunicipalCode          string               `json:"f_kallikratis_pararthmatos"`
	Comments               string               `json:"f_comments"`
	LegalRepTaxID          string               `json:"f_afm_proswpoy"`
	Employees              OvertimeEmployeeList `json:"Ergazomenoi"`
}

// OvertimeEmployeeList holds the overtime of the employees of a branch.
type OvertimeEmployeeList struct {
	Records []OvertimeRecord `json:"OvertimeErgazomenosDate"`
}

// OvertimeRecord is the overtime of one employee on one day.
//
// Every field is a string. The second interval is optional and is omitted
// when it is not used.
type OvertimeRecord struct {
	EmployeeTaxID     string `json:"f_afm"`
	EmployeeSSN       string `json:"f_amka"`
	EmployeeLastName  string `json:"f_eponymo"`
	EmployeeFirstName string `json:"f_onoma"`
	Date              Date   `json:"f_date"`
	FromTime          Time   `json:"f_from"`
	ToTime            Time   `json:"f_to"`
	SecondFromTime    *Time  `json:"f_from_2,omitempty"`
	SecondToTime      *Time  `json:"f_to_2,omitempty"`
	Cancelled         Bool   `json:"f_cancellation"`
	ProfessionCode    string `json:"f_step"`
	Justification     string `json:"f_reason"`
	WeekPattern       string `json:"f_weekdates"`
	ASEEApproval      string `json:"f_asee"`
}
