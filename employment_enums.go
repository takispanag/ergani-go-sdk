package ergani

// The enumerated fields of the employment lifecycle declarations. The codes
// are the ones the API documents, and the official description of every field
// is in the documentation of the interoperability guide of the Ministry.

// Sex is the sex of the employee.
type Sex string

// The sexes the API accepts.
const (
	// SexMale is a male employee.
	SexMale Sex = "0"

	// SexFemale is a female employee.
	SexFemale Sex = "1"
)

// MaritalStatus is the marital status of the employee.
type MaritalStatus string

// The marital statuses the API accepts.
const (
	// MaritalSingle is an unmarried employee.
	MaritalSingle MaritalStatus = "0"

	// MaritalMarried is a married employee.
	MaritalMarried MaritalStatus = "1"

	// MaritalDivorced is a divorced employee.
	MaritalDivorced MaritalStatus = "2"

	// MaritalWidowed is a widowed employee.
	MaritalWidowed MaritalStatus = "3"
)

// WorkerCategory is how the employee is characterised towards the employment
// relationship.
type WorkerCategory string

// The categories the API accepts.
const (
	// WorkerCategoryWorker is an employee characterised as a worker.
	WorkerCategoryWorker WorkerCategory = "0"

	// WorkerCategoryEmployee is an employee characterised as an employee.
	WorkerCategoryEmployee WorkerCategory = "1"
)

// RelationshipType is the kind of the employment relationship.
type RelationshipType string

// The relationships the API accepts.
const (
	// RelationshipIndefinite is an employment of an indefinite duration.
	RelationshipIndefinite RelationshipType = "0"

	// RelationshipFixedTerm is an employment of a fixed duration.
	RelationshipFixedTerm RelationshipType = "1"

	// RelationshipProject is an employment for a project.
	RelationshipProject RelationshipType = "2"
)

// EmploymentStatus is how much of the working time the employment covers.
type EmploymentStatus string

// The statuses the API accepts.
//
// The API also knows a fourth value, the employment on demand, but the
// schemas that are active reject it.
const (
	// EmploymentFull is a full time employment.
	EmploymentFull EmploymentStatus = "0"

	// EmploymentPartial is a part time employment.
	EmploymentPartial EmploymentStatus = "1"

	// EmploymentRotation is an employment that rotates.
	EmploymentRotation EmploymentStatus = "2"
)

// EssentialTermsAcceptance is how the employee accepted the essential terms
// of the employment.
type EssentialTermsAcceptance string

// The acceptances the API accepts.
const (
	// EssentialTermsWithFile is an acceptance with the signed document.
	EssentialTermsWithFile EssentialTermsAcceptance = "0"

	// EssentialTermsInMyErgani is an acceptance recorded in myErgani.
	EssentialTermsInMyErgani EssentialTermsAcceptance = "1"
)

// IndividualAgreementStatus is the state of the individual agreement of the
// employee.
type IndividualAgreementStatus string

// The states the API accepts.
const (
	// IndividualAgreementNone is an employment without an individual
	// agreement.
	IndividualAgreementNone IndividualAgreementStatus = "0"

	// IndividualAgreementAccepted is an employment with an accepted individual
	// agreement.
	IndividualAgreementAccepted IndividualAgreementStatus = "1"

	// IndividualAgreementPending is an employment whose individual agreement
	// is pending.
	IndividualAgreementPending IndividualAgreementStatus = "2"
)
