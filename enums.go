package ergani

// MovementType is the kind of a work card movement.
type MovementType string

// The movement types of a work card.
const (
	// MovementArrival reports that the employee arrived at work.
	MovementArrival MovementType = "ARRIVAL"

	// MovementDeparture reports that the employee left work.
	MovementDeparture MovementType = "DEPARTURE"
)

// Valid reports whether the movement type is one the API accepts.
func (t MovementType) Valid() bool {
	switch t {
	case MovementArrival, MovementDeparture:
		return true
	default:
		return false
	}
}

// code returns the code the API uses for the movement type.
func (t MovementType) code() (string, bool) {
	switch t {
	case MovementArrival:
		return "0", true
	case MovementDeparture:
		return "1", true
	default:
		return "", false
	}
}

// LateJustification explains why a work card is declared late.
type LateJustification string

// The reasons of a late declaration.
//
// The API publishes the list of the reasons at runtime, under the
// WorkCardDelayReason parameter of the code list service.
const (
	// JustificationPowerOutage reports an interruption of the electricity.
	JustificationPowerOutage LateJustification = "POWER_OUTAGE"

	// JustificationEmployerSystemsUnavailable reports a failure of the
	// systems of the employer.
	JustificationEmployerSystemsUnavailable LateJustification = "EMPLOYER_SYSTEMS_UNAVAILABLE"

	// JustificationErganiSystemsUnavailable reports a failure of the systems
	// of Ergani.
	JustificationErganiSystemsUnavailable LateJustification = "ERGANI_SYSTEMS_UNAVAILABLE"
)

// Valid reports whether the justification is one the API accepts.
func (j LateJustification) Valid() bool {
	switch j {
	case JustificationPowerOutage,
		JustificationEmployerSystemsUnavailable,
		JustificationErganiSystemsUnavailable:
		return true
	default:
		return false
	}
}

// code returns the code the API uses for the justification.
func (j LateJustification) code() (string, bool) {
	switch j {
	case JustificationPowerOutage:
		return "001", true
	case JustificationEmployerSystemsUnavailable:
		return "002", true
	case JustificationErganiSystemsUnavailable:
		return "003", true
	default:
		return "", false
	}
}

// WorkType is the kind of a working period.
//
// The codes are the ones the API uses for the f_type field of a working time
// analysis. They are the same codes the platform sends in the SMS messages of
// the individual working time declaration.
type WorkType string

// The working time types.
const (
	// WorkTypeOffice is a working period at the premises of the employer.
	WorkTypeOffice WorkType = "OFFICE"

	// WorkTypeRemote is a working period performed remotely.
	WorkTypeRemote WorkType = "REMOTE"

	// WorkTypeRestDay is a rest day.
	WorkTypeRestDay WorkType = "REST_DAY"

	// WorkTypeNoWork is a period without work, used by the employees who
	// rotate.
	WorkTypeNoWork WorkType = "NO_WORK"

	// WorkTypeOvertime is an overtime period.
	WorkTypeOvertime WorkType = "OVERTIME"

	// WorkTypeNoOvertime cancels an overtime period.
	WorkTypeNoOvertime WorkType = "NO_OVERTIME"
)

// Valid reports whether the working time type is one the API accepts.
func (t WorkType) Valid() bool {
	switch t {
	case WorkTypeOffice, WorkTypeRemote, WorkTypeRestDay, WorkTypeNoWork,
		WorkTypeOvertime, WorkTypeNoOvertime:
		return true
	default:
		return false
	}
}

// code returns the code the API uses for the working time type.
func (t WorkType) code() (string, bool) {
	switch t {
	case WorkTypeOffice:
		return "ΕΡΓ", true
	case WorkTypeRemote:
		return "ΤΗΛ", true
	case WorkTypeRestDay:
		return "ΑΝ", true
	case WorkTypeNoWork:
		return "ΜΕ", true
	case WorkTypeOvertime:
		return "ΥΠ", true
	case WorkTypeNoOvertime:
		return "ΧΥΠ", true
	default:
		return "", false
	}
}

// LeaveType is the kind of a leave.
//
// The codes are the ones the API uses for the f_type field of a leave
// declaration. The list is the one of the official interoperability guide.
type LeaveType string

// The leave types.
const (
	LeaveRegular          LeaveType = "REGULAR"
	LeaveBloodDonation    LeaveType = "BLOOD_DONATION"
	LeaveExamination      LeaveType = "EXAMINATION"
	LeaveUnpaid           LeaveType = "UNPAID"
	LeaveMaternity        LeaveType = "MATERNITY"
	LeaveMaternitySpecial LeaveType = "MATERNITY_SPECIAL"
	LeavePaternity        LeaveType = "PATERNITY"
	LeaveChildCare        LeaveType = "CHILD_CARE"
	LeaveParental         LeaveType = "PARENTAL"
	LeaveCarer            LeaveType = "CARER"
	LeaveForceMajeure     LeaveType = "FORCE_MAJEURE"
	LeaveAssistedReprod   LeaveType = "ASSISTED_REPRODUCTION"
	LeavePrenatalExam     LeaveType = "PRENATAL_EXAMINATION"
	LeaveMarriage         LeaveType = "MARRIAGE"
	LeaveSeriousIllness   LeaveType = "SERIOUS_ILLNESS"
	LeaveChildHospital    LeaveType = "CHILD_HOSPITALISATION"
	LeaveSingleParent     LeaveType = "SINGLE_PARENT"
	LeaveSchoolMonitoring LeaveType = "SCHOOL_MONITORING"
	LeaveDependentIllness LeaveType = "DEPENDENT_ILLNESS"
	LeaveDangerOfViolence LeaveType = "DANGER_OF_VIOLENCE"
	LeaveSick             LeaveType = "SICK"
	LeaveDisabled         LeaveType = "DISABLED"
	LeaveBereavement      LeaveType = "BEREAVEMENT"
	LeaveMinorStudent     LeaveType = "MINOR_STUDENT"
	LeaveBloodTransfusion LeaveType = "BLOOD_TRANSFUSION"
	LeaveTraining         LeaveType = "TRAINING"
	LeaveAIDS             LeaveType = "AIDS"
	LeaveFlexibleWork     LeaveType = "FLEXIBLE_WORK"
	LeaveChildCareHours   LeaveType = "CHILD_CARE_HOURS"
	LeaveParentalHours    LeaveType = "PARENTAL_HOURS"
	LeaveForceMajeureHour LeaveType = "FORCE_MAJEURE_HOURS"
	LeaveFlexibleHours    LeaveType = "FLEXIBLE_HOURS"
	LeavePrenatalHours    LeaveType = "PRENATAL_HOURS"
	LeaveSchoolHours      LeaveType = "SCHOOL_HOURS"
	LeaveOther            LeaveType = "OTHER"
	LeaveOtherHours       LeaveType = "OTHER_HOURS"
)

// Valid reports whether the leave type is one the API accepts.
func (t LeaveType) Valid() bool {
	_, ok := t.code()
	return ok
}

// code returns the code the API uses for the leave type.
func (t LeaveType) code() (string, bool) {
	switch t {
	case LeaveRegular:
		return "ΑΔΚΑΝ", true
	case LeaveBloodDonation:
		return "ΑΔΑΙΜ", true
	case LeaveExamination:
		return "ΑΔΕΞ", true
	case LeaveUnpaid:
		return "ΑΔΑΑ", true
	case LeaveMaternity:
		return "ΑΔΜΗ", true
	case LeaveMaternitySpecial:
		return "ΑΔΠΠΜ", true
	case LeavePaternity:
		return "ΑΔΠΑ", true
	case LeaveChildCare:
		return "ΑΔΦΠ", true
	case LeaveParental:
		return "ΑΔΓΟΝ", true
	case LeaveCarer:
		return "ΑΔΦΡΟ", true
	case LeaveForceMajeure:
		return "ΑΔΑΠΑΒ", true
	case LeaveAssistedReprod:
		return "ΑΔΙΥΑ", true
	case LeavePrenatalExam:
		return "ΑΔΠΕ", true
	case LeaveMarriage:
		return "ΑΔΓΑΜ", true
	case LeaveSeriousIllness:
		return "ΑΔΣΝΠ", true
	case LeaveChildHospital:
		return "ΑΔΝΠ", true
	case LeaveSingleParent:
		return "ΑΔΜΟ", true
	case LeaveSchoolMonitoring:
		return "ΑΔΠΣΕΤ", true
	case LeaveDependentIllness:
		return "ΑΔΑΠΕΜ", true
	case LeaveDangerOfViolence:
		return "ΑΔΑΠΣΚ", true
	case LeaveSick:
		return "ΑΔΑΣ", true
	case LeaveDisabled:
		return "ΑΔΑΜΕΑ", true
	case LeaveBereavement:
		return "ΑΔΘΣΥΓ", true
	case LeaveMinorStudent:
		return "ΑΔΑΝΣΠ", true
	case LeaveBloodTransfusion:
		return "ΑΔΜΑΑ", true
	case LeaveTraining:
		return "ΑΔΕΚΦ", true
	case LeaveAIDS:
		return "ΑΔΣΕΑΑ", true
	case LeaveFlexibleWork:
		return "ΑΔΕΡΕ", true
	case LeaveChildCareHours:
		return "ΩΑΦΠ", true
	case LeaveParentalHours:
		return "ΩΑΓΟΝ", true
	case LeaveForceMajeureHour:
		return "ΩΑΑΠΑΒ", true
	case LeaveFlexibleHours:
		return "ΩΑΕΡΕ", true
	case LeavePrenatalHours:
		return "ΩΑΠΕ", true
	case LeaveSchoolHours:
		return "ΩΑΠΣΕΤ", true
	case LeaveOther:
		return "ΑΔΑΛ", true
	case LeaveOtherHours:
		return "ΩΑΑΛ", true
	default:
		return "", false
	}
}

// OvertimeJustification is the official reason for overtime.
type OvertimeJustification string

// The justifications of an overtime declaration.
//
// The API publishes the list at runtime, under the WorkTimeType parameter of
// the code list service.
const (
	// OvertimeAccidentPrevention prevents accidents or restores damage.
	OvertimeAccidentPrevention OvertimeJustification = "ACCIDENT_PREVENTION_OR_DAMAGE_RESTORATION"

	// OvertimeUrgentSeasonal covers urgent seasonal work.
	OvertimeUrgentSeasonal OvertimeJustification = "URGENT_SEASONAL_TASKS"

	// OvertimeExceptionalWorkload covers an exceptional workload.
	OvertimeExceptionalWorkload OvertimeJustification = "EXCEPTIONAL_WORKLOAD"

	// OvertimeSupplementary covers supplementary work.
	OvertimeSupplementary OvertimeJustification = "SUPPLEMENTARY_TASKS"

	// OvertimeLostHoursSudden covers hours lost to sudden causes.
	OvertimeLostHoursSudden OvertimeJustification = "LOST_HOURS_SUDDEN_CAUSES"

	// OvertimeLostHoursHolidays covers hours lost to official holidays.
	OvertimeLostHoursHolidays OvertimeJustification = "LOST_HOURS_OFFICIAL_HOLIDAYS"

	// OvertimeLostHoursWeather covers hours lost to the weather.
	OvertimeLostHoursWeather OvertimeJustification = "LOST_HOURS_WEATHER_CONDITIONS"

	// OvertimeEmergencyClosure covers an emergency closure of the day.
	OvertimeEmergencyClosure OvertimeJustification = "EMERGENCY_CLOSURE_DAY"

	// OvertimeNonWorkday covers work on a day that is not a working day.
	OvertimeNonWorkday OvertimeJustification = "NON_WORKDAY_TASKS"
)

// Valid reports whether the justification is one the API accepts.
func (j OvertimeJustification) Valid() bool {
	_, ok := j.code()
	return ok
}

// code returns the code the API uses for the justification.
func (j OvertimeJustification) code() (string, bool) {
	switch j {
	case OvertimeAccidentPrevention:
		return "001", true
	case OvertimeUrgentSeasonal:
		return "002", true
	case OvertimeExceptionalWorkload:
		return "003", true
	case OvertimeSupplementary:
		return "004", true
	case OvertimeLostHoursSudden:
		return "005", true
	case OvertimeLostHoursHolidays:
		return "006", true
	case OvertimeLostHoursWeather:
		return "007", true
	case OvertimeEmergencyClosure:
		return "008", true
	case OvertimeNonWorkday:
		return "009", true
	default:
		return "", false
	}
}

// WeekPattern is the number of working days of a week.
type WeekPattern string

// The weekly working patterns.
const (
	// WeekFiveDays is a five day week.
	WeekFiveDays WeekPattern = "FIVE_DAYS"

	// WeekSixDays is a six day week.
	WeekSixDays WeekPattern = "SIX_DAYS"
)

// Valid reports whether the weekly pattern is one the API accepts.
func (p WeekPattern) Valid() bool {
	switch p {
	case WeekFiveDays, WeekSixDays:
		return true
	default:
		return false
	}
}

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
