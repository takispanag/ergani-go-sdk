package wire

// WorkCardDeclarations is the root of a work card submission.
//
// The schema is Card_v1.xsd. The root holds the cards of every business
// branch of the employer.
type WorkCardDeclarations struct {
	Cards WorkCardList `json:"Cards"`
}

// WorkCardList holds the work cards of the submission.
type WorkCardList struct {
	Card []CompanyWorkCard `json:"Card"`
}

// CompanyWorkCard groups the work cards of one business branch.
//
// The schema nests the movements of the employees one level deeper, under
// "Details".
type CompanyWorkCard struct {
	EmployerTaxID string          `json:"f_afm_ergodoti"`
	BranchNumber  string          `json:"f_aa"`
	Comments      string          `json:"f_comments"`
	Details       WorkCardDetails `json:"Details"`
}

// WorkCardDetails holds the movements of the employees of a branch.
type WorkCardDetails struct {
	CardDetails []WorkCard `json:"CardDetails"`
}

// WorkCard is one movement of one employee.
//
// The schema declares every field as a string, and the reference date as an
// xs:date.
type WorkCard struct {
	EmployeeTaxID     string   `json:"f_afm"`
	EmployeeLastName  string   `json:"f_eponymo"`
	EmployeeFirstName string   `json:"f_onoma"`
	MovementType      string   `json:"f_type"`
	ReferenceDate     ISODate  `json:"f_reference_date"`
	MovementDateTime  DateTime `json:"f_date"`
	// LateJustification explains why the movement is declared late. The API
	// expects null when there is none.
	LateJustification *string `json:"f_aitiologia"`
}
