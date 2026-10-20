package ergani

// DocumentCode identifies a type of declaration.
//
// A code is the last segment of the path of a submission, as in
// "POST /Documents/WRKCardSE". The set of codes an employer may use depends
// on the account: [Client.ListSubmissions] returns the codes that are active
// for the connected user.
//
// The API adds codes over time. A code that is not listed here can still be
// used with [Client.Submit].
type DocumentCode string

// The declaration types of the Ergani platform.
//
// The card sector, which is the regime of the digital work card.
const (
	// CodeWorkCard is the declaration of the start and the end of the
	// employment of an employee.
	CodeWorkCard DocumentCode = "WRKCardSE"

	// CodeWorkStatusChange is the declaration of a change of the employment
	// relationship, for the working time organisation.
	CodeWorkStatusChange DocumentCode = "WKChgWK"

	// CodeScheduleDaily is the working time declaration that varies from day
	// to day.
	CodeScheduleDaily DocumentCode = "WTODaily"

	// CodeScheduleDailyAccounting is the retrospective variant of the
	// variable daily working time declaration.
	CodeScheduleDailyAccounting DocumentCode = "WTODailyA"

	// CodeScheduleDailyDrivers is the variant of the variable daily working
	// time declaration for the employees who drive.
	CodeScheduleDailyDrivers DocumentCode = "WTODailyD"

	// CodeScheduleWeekly is the working time declaration of a fixed weekly
	// schedule.
	CodeScheduleWeekly DocumentCode = "WTOWeek"

	// CodeScheduleOvertime is the working time declaration of overtime.
	CodeScheduleOvertime DocumentCode = "WTOOv"

	// CodeScheduleOvertimeAccounting is the retrospective variant of the
	// overtime working time declaration.
	CodeScheduleOvertimeAccounting DocumentCode = "WTOOvA"

	// CodeScheduleOvertimeDrivers is the variant of the overtime working time
	// declaration for the employees who drive.
	CodeScheduleOvertimeDrivers DocumentCode = "WTOOvD"

	// CodeLeave is the working time declaration of a leave.
	//
	// The Ministry does not publish the code of the leave declaration: the
	// catalogue of the official interoperability guide leaves the cell
	// empty, and the sample archive names the file of the example
	// "wtoHoliday". The code below is the one the reference implementations
	// use. [Client.ListSubmissions] reports the code of the connected
	// employer, and it is the authoritative source.
	CodeLeave DocumentCode = "WTOLeave"

	// CodeLeaveCorrection is the correction of a leave declaration.
	CodeLeaveCorrection DocumentCode = "WTOLeaveC"

	// CodeExemptionNotification is the declaration of an exemption from the
	// duty of advance notification.
	CodeExemptionNotification DocumentCode = "ExProan"

	// CodeSixthDay is the declaration of the employment on an emergency
	// shift.
	CodeSixthDay DocumentCode = "SixthDay"
)

// The declarations of the start of the employment.
const (
	// CodeE3 is the unified E3 hiring notification.
	CodeE3 DocumentCode = "E3"

	// CodeHire is the digital notification of the start of an employment.
	CodeHire DocumentCode = "WebE3N"

	// CodeHireTransfer is the notification of the start of an employment
	// with a transfer from another company.
	CodeHireTransfer DocumentCode = "WebE3M"

	// CodeHireBorrowed is the notification of the start of an employment with
	// staff borrowed from another company.
	CodeHireBorrowed DocumentCode = "WebE3D"

	// CodeHireForLoanOut is the notification of the hire of an employee who
	// is loaned out to another company.
	CodeHireForLoanOut DocumentCode = "WebE3PD"

	// CodeHireEmergency is the notification of the hire of an employee to
	// cover an emergency need.
	CodeHireEmergency DocumentCode = "FastAna"

	// CodeHireEmergencyCorrection is the correction of an emergency hire.
	CodeHireEmergencyCorrection DocumentCode = "FastAnaC"

	// CodeEmploymentChange is the notification of a change of the details of
	// an employment.
	CodeEmploymentChange DocumentCode = "WebMA"

	// CodeEmploymentChangeBorrowed is the notification of a change of the
	// details of the employment of borrowed staff.
	CodeEmploymentChangeBorrowed DocumentCode = "WebMAD"
)

// The declarations of the end of the employment.
const (
	// CodeResignation is the declaration of a voluntary resignation.
	CodeResignation DocumentCode = "WebE5N"

	// CodeResignationNotice is the declaration of the notice of resignation.
	CodeResignationNotice DocumentCode = "WebE5O"

	// CodeResignationAfterNotice is the declaration of a resignation that
	// follows a notice.
	CodeResignationAfterNotice DocumentCode = "WebE5AO"

	// CodeVoluntaryExit is the declaration of a voluntary exit.
	CodeVoluntaryExit DocumentCode = "WebE5E"

	// CodeRetirementByResignation is the declaration of a retirement.
	CodeRetirementByResignation DocumentCode = "WebE5S"

	// CodeResignationFifteenYears is the declaration of a resignation after
	// fifteen years of service or upon reaching the retirement age.
	CodeResignationFifteenYears DocumentCode = "WebE5DS"

	// CodeDeath is the declaration of the end of an employment because of the
	// death of the employee.
	CodeDeath DocumentCode = "WebE5D"

	// CodeDismissalWithoutNotice is the declaration of a dismissal without
	// notice.
	CodeDismissalWithoutNotice DocumentCode = "WebE6NXP"

	// CodeDismissalWithNotice is the declaration of a dismissal with notice.
	CodeDismissalWithNotice DocumentCode = "WebE6NMP"

	// CodeRetirementByDismissal is the declaration of a retirement by
	// dismissal without notice.
	CodeRetirementByDismissal DocumentCode = "WebE6SXP"

	// CodeProbationTermination is the declaration of the termination of a
	// probation period.
	CodeProbationTermination DocumentCode = "WebE6LT"

	// CodeTransferOut is the declaration of the transfer of an employee to
	// another company.
	CodeTransferOut DocumentCode = "WebE6M"

	// CodeLoanOutEnd is the declaration of the end of a loan-out.
	CodeLoanOutEnd DocumentCode = "WebE6LD"

	// CodeFixedTermTermination is the declaration of the termination of a
	// fixed term contract.
	CodeFixedTermTermination DocumentCode = "WebE7N"
)

// The declarations of the special categories of workers and the reports.
const (
	// CodeConstructionWorker is the notification of a worker on a
	// construction site.
	CodeConstructionWorker DocumentCode = "WebE12"

	// CodeConstructionWorkerCensus is the census notification of a worker on
	// a construction site.
	CodeConstructionWorkerCensus DocumentCode = "WebE12Ap"

	// CodeParentalLeaveGrant is the declaration of the employer for the
	// parental leave grant.
	CodeParentalLeaveGrant DocumentCode = "WebE14"

	// CodeParentalLeaveGrantCorrection is the correction of the parental
	// leave grant declaration.
	CodeParentalLeaveGrantCorrection DocumentCode = "WebE14C"

	// CodeAnnualLeaveNotification is the notification of the annual regular
	// leave entitlement.
	CodeAnnualLeaveNotification DocumentCode = "WebEA"

	// CodeIGEEReport is the IGEE reconciling report.
	CodeIGEEReport DocumentCode = "WebIgee"

	// CodeEPAReport is the EPA reconciling report.
	CodeEPAReport DocumentCode = "WebEpa"

	// CodeVehicleUse is the declaration of the employees that use a vehicle
	// or a motorcycle.
	CodeVehicleUse DocumentCode = "WebE13"

	// CodeInternshipStart is the declaration of the start of an internship.
	CodeInternshipStart DocumentCode = "WebE35"

	// CodeInternshipStartCorrection is the correction of the declaration of
	// the start of an internship.
	CodeInternshipStartCorrection DocumentCode = "WebE35C"

	// CodeInternshipEnd is the declaration of the end of an internship.
	CodeInternshipEnd DocumentCode = "WebE35D"

	// CodeInternshipScheduleChange is the declaration of a change of the
	// schedule of an internship.
	CodeInternshipScheduleChange DocumentCode = "WebE35T"

	// CodeBeneficiaryEmployment is the declaration of the employment of a
	// beneficiary of a public benefit programme.
	CodeBeneficiaryEmployment DocumentCode = "WebE31"

	// CodeBeneficiaryEmploymentCorrection is the correction of the
	// declaration of the employment of a beneficiary.
	CodeBeneficiaryEmploymentCorrection DocumentCode = "WebE31C"

	// CodeBeneficiaryEmploymentExtension is the declaration of the extension
	// of the employment of a beneficiary.
	CodeBeneficiaryEmploymentExtension DocumentCode = "WebE31E"

	// CodeBeneficiaryEmploymentSuspension is the declaration of the
	// suspension of the employment of a beneficiary.
	CodeBeneficiaryEmploymentSuspension DocumentCode = "WebE31D"

	// CodeTheoreticalTraining is the declaration of the start of a
	// theoretical training.
	CodeTheoreticalTraining DocumentCode = "WebE32"

	// CodeTheoreticalTrainingCorrection is the correction of the declaration
	// of the start of a theoretical training.
	CodeTheoreticalTrainingCorrection DocumentCode = "WebE32C"

	// CodeTerminationE4 is the E4 declaration of the end of an employment.
	//
	// It was replaced by the digital termination announcements of the E5,
	// E6 and E7 families.
	CodeTerminationE4 DocumentCode = "E4"

	// CodeOvertimeE8 is the E8 declaration of legal overtime.
	//
	// The E8 form was replaced by the working time overtime declarations.
	// The API still answers the route, but it rejects the payload of the
	// form: the overtime of a branch is declared with
	// [CodeScheduleOvertime] and its variants.
	CodeOvertimeE8 DocumentCode = "OvTime"
)
