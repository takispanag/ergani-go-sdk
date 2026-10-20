package main

// The tables that drive the generator.
//
// One table holds the fields the public API names itself, one holds the
// families of declarations, and the rest hold the way a kind of field is
// written in Go and on the wire. Everything else is read from the schemas
// and the reference payloads of the Ministry.

// knownField is the kind of a field and the name it carries in the public
// API. Every other field is inferred from the schema of its declaration.
type knownField struct {
	kind string
	name string
}

// knownFields holds the fields the public API names itself.
var knownFields = map[string]knownField{
	"f_aa_pararthmatos":                       {"string", "BranchNumber"},
	"f_rel_protocol":                          {"string", "RelatedProtocol"},
	"f_rel_date":                              {"date", "RelatedDate"},
	"f_ypiresia_sepe":                         {"string", "SEPEServiceCode"},
	"f_ypiresia_oaed":                         {"string", "OAEEDServiceCode"},
	"f_kad_pararthmatos":                      {"string", "BranchActivityCode"},
	"f_kallikratis_pararthmatos":              {"string", "MunicipalCode"},
	"f_eponymo":                               {"string", "LastName"},
	"f_onoma":                                 {"string", "FirstName"},
	"f_onoma_patros":                          {"string", "FatherName"},
	"f_onoma_mitros":                          {"string", "MotherName"},
	"f_birthdate":                             {"date", "BirthDate"},
	"f_sex":                                   {"enum:Sex", "Sex"},
	"f_yphkoothta":                            {"string", "NationalityCode"},
	"f_typos_taytothtas":                      {"string", "IdentityDocumentType"},
	"f_ar_taytothtas":                         {"string", "IdentityDocumentNumber"},
	"f_ekdousa_arxh":                          {"string", "IdentityDocumentIssuer"},
	"f_date_ekdosis":                          {"date", "IdentityDocumentIssued"},
	"f_date_ekdosis_lixi":                     {"date", "IdentityDocumentExpires"},
	"f_res_permit_inst":                       {"flag", "HasResidencePermit"},
	"f_res_permit_inst_type":                  {"string", "ResidencePermitType"},
	"f_res_permit_inst_ar":                    {"string", "ResidencePermitNumber"},
	"f_res_permit_inst_lixi":                  {"date", "ResidencePermitExpires"},
	"f_res_permit_ap":                         {"flag", "HasWorkPermit"},
	"f_res_permit_ap_type":                    {"string", "WorkPermitType"},
	"f_res_permit_ap_ar":                      {"string", "WorkPermitNumber"},
	"f_res_permit_ap_lixi":                    {"date", "WorkPermitExpires"},
	"f_res_permit_visa":                       {"flag", "HasVisa"},
	"f_res_permit_visa_ar":                    {"string", "VisaNumber"},
	"f_res_permit_visa_from":                  {"date", "VisaFrom"},
	"f_res_permit_visa_to":                    {"date", "VisaTo"},
	"f_marital_status":                        {"enum:MaritalStatus", "MaritalStatus"},
	"f_arithmos_teknon":                       {"count", "Children"},
	"f_afm":                                   {"string", "TaxID"},
	"f_doy":                                   {"string", "TaxOffice"},
	"f_amika":                                 {"string", "SocialSecurityRegistryNumber"},
	"f_amka":                                  {"string", "SocialSecurityNumber"},
	"f_code_anergias":                         {"string", "UnemploymentCode"},
	"f_ar_vivliou_anilikou":                   {"string", "AnnualIncome"},
	"f_epipedo_morfosis":                      {"string", "EducationLevel"},
	"f_xaraktirismos":                         {"enum:WorkerCategory", "WorkerCategory"},
	"f_sxeshapasxolisis":                      {"enum:RelationshipType", "RelationshipType"},
	"f_orismenou_apo":                         {"date", "FixedTermFrom"},
	"f_orismenou_ews":                         {"date", "FixedTermTo"},
	"f_proslipsidate":                         {"date", "HireDate"},
	"f_proslipsitime":                         {"time", "HireTime"},
	"f_apoxwrisidate":                         {"date", "LeaveDate"},
	"f_apoxwrisitime":                         {"time", "LeaveTime"},
	"f_eidikothta":                            {"string", "ProfessionCode"},
	"f_eidikothta_anal":                       {"string", "ProfessionDescription"},
	"f_proipiresia":                           {"string", "InsuranceFund"},
	"f_apodoxes":                              {"amount", "MonthlyWages"},
	"f_hour_apodoxes":                         {"amount", "HourlyWages"},
	"f_kathestosapasxolisis":                  {"enum:EmploymentStatus", "EmploymentStatus"},
	"f_week_hours":                            {"hours", "WeeklyHours"},
	"f_full_employment_hours":                 {"hours", "FullEmploymentHours"},
	"f_week_days":                             {"enum:WeekPattern", "WeeklyPattern"},
	"f_working_card":                          {"flag", "UsesWorkCard"},
	"f_working_time_digital_organization":     {"flag", "UsesDigitalWorkingTime"},
	"f_dialeimma_minutes":                     {"minutes", "BreakMinutes"},
	"f_dialeimma_entos_wrariou":               {"flag", "BreakWithinSchedule"},
	"f_euelikto_wrario_minutes":               {"minutes", "FlexibleArrivalMinutes"},
	"f_responsible_position":                  {"string", "ResponsiblePosition"},
	"f_kyria_asfalisi":                        {"string", "PrimaryInsuranceFund"},
	"f_prosthetes_asfalistikes":               {"string", "SupplementaryInsuranceFund"},
	"f_ipoxreotiki_katartisi":                 {"string", "RehabilitationRecord"},
	"f_efarmoste_sillogiki_simbasi":           {"flag", "HasCollectiveAgreement"},
	"f_efarmoste_sillogiki_simbasi_comments":  {"string", "CollectiveAgreementNotes"},
	"f_efarmostea_sillogiki_simbasi":          {"flag", "HasSignedCollectiveAgreement"},
	"f_efarmostea_sillogiki_simbasi_comments": {"string", "SignedCollectiveAgreementNotes"},
	"f_mh_provlepsimo_programma":              {"flag", "HasUnlistedProgram"},
	"f_topos_ergasias":                        {"flag", "HasDeclaredPlaceOfWork"},
	"f_topos_ergasias_comment":                {"string", "PlaceOfWorkNotes"},
	"f_paraggelia_hmeres_hours":               {"string", "NoticeDaysAndHours"},
	"f_paraggelia_min_notification":           {"string", "MinimumNoticePeriod"},
	"f_paraggelia_notes":                      {"string", "NoticeNotes"},
	"f_xronos_katavolis_apodoxon":             {"string", "WagesPaymentFrequency"},
	"f_comments":                              {"string", "Comments"},
	"f_file":                                  {"base64", "File"},
	"f_foreign_file":                          {"base64", "ForeignWorkerFile"},
	"f_young_file":                            {"base64", "MinorWorkerFile"},
	"f_file_symbash":                          {"base64", "IndividualAgreementFile"},
	"f_basics_acceptance":                     {"enum:EssentialTermsAcceptance", "EssentialTermsAcceptance"},
	"f_atomikh_symbash":                       {"enum:IndividualAgreementStatus", "IndividualAgreementStatus"},
	"f_special_case":                          {"string", "SpecialCase"},
	"f_trial_period":                          {"flag", "HasTrialPeriod"},
	"f_trial_date_to":                         {"date", "TrialPeriodEnds"},
	"f_topothetisioaed":                       {"flag", "HasPriorProgramPlacement"},
	"f_programaoaed":                          {"string", "ProgramPlacement"},
	"f_replaceprograma":                       {"flag", "ReplacesBeneficiary"},
	"f_replaceprograma_afm":                   {"string", "ReplacedBeneficiaryTaxID"},
	"f_replaceprograma_amka":                  {"string", "ReplacedBeneficiarySocialSecurityNumber"},
	"f_transfer_company_afm":                  {"string", "TransferCompanyTaxID"},
	"f_transfer_company_eponimia":             {"string", "TransferCompanyName"},
	"f_date_metabibashs":                      {"date", "TransferDate"},
	"f_borrow_company_afm":                    {"string", "BorrowingCompanyTaxID"},
	"f_borrow_company_eponimia":               {"string", "BorrowingCompanyName"},
	"f_borrow_date_from":                      {"date", "LoanFrom"},
	"f_borrow_date_to":                        {"date", "LoanTo"},
	"f_borrow_type":                           {"string", "LoanType"},
	"f_ergazom_borrow_agreement":              {"string", "BorrowedStaffAgreement"},
	"f_kataboli_apodoxon":                     {"string", "WagesPaymentMethod"},
	"f_apolysisdate":                          {"date", "TerminationDate"},
	"f_lixisymbashdate":                       {"date", "ContractEndDate"},
	"f_logosperatosis":                        {"string", "TerminationReason"},
	"f_logosperatosiscomments":                {"string", "TerminationReasonNotes"},
	"f_oros":                                  {"string", "Month"},
	"f_omadiki":                               {"flag", "HasCollectiveAction"},
	"f_omadikidate":                           {"date", "CollectiveActionDate"},
	"f_omadikiarithmos":                       {"string", "CollectiveActionNumber"},
	"f_minesproidopoihsh":                     {"string", "NotificationMonths"},
	"f_proidopoihshdate":                      {"date", "NoticeDate"},
	"f_koinopoihshdate":                       {"date", "DismissalDate"},
	"f_posoapozimiosis":                       {"amount", "DismissalCompensation"},
	"f_tropoi_oxlhshs":                        {"string", "ResignationMethods"},
	"f_oxlhsh_date_ypovolis":                  {"date", "NotificationServiceDate"},
	"f_oxlhsh_protocol":                       {"string", "NotificationServiceProtocol"},
}

// initialisms holds the abbreviations the public API spells in capitals.
var initialisms = map[string]string{
	"afm": "AFM", "amka": "AMKA", "amika": "AMIKA", "sepe": "SEPE",
	"oaed": "OAED", "kad": "KAD", "id": "ID", "url": "URL", "pdf": "PDF",
	"vat": "VAT", "iban": "IBAN", "covid": "COVID", "eae": "EAE",
}

// semantic holds the name of the constant of a variant, when the code alone
// would not say what the declaration is.
var semantic = map[string]string{
	"WebE3N":   "HireNew",
	"WebE3M":   "HireTransfer",
	"WebE3D":   "HireBorrowed",
	"WebE3PD":  "HireForLoanOut",
	"WebE5N":   "TerminationResignation",
	"WebE5O":   "TerminationNotice",
	"WebE5AO":  "TerminationResignationAfterNotice",
	"WebE5E":   "TerminationVoluntaryExit",
	"WebE5S":   "TerminationRetirement",
	"WebE5DS":  "TerminationFifteenYears",
	"WebE5D":   "TerminationDeath",
	"WebE6NXP": "TerminationDismissalWithoutNotice",
	"WebE6NMP": "TerminationDismissalWithNotice",
	"WebE6SXP": "TerminationRetirementByDismissal",
	"WebE6LT":  "TerminationProbation",
	"WebE6M":   "TerminationTransfer",
	"WebE7N":   "TerminationFixedTerm",
	"WebMA":    "ChangeEmployment",
	"WebMAD":   "ChangeBorrowedEmployment",
	"WebE12":   "ConstructionWorker",
	"WebE12Ap": "ConstructionWorkerCensus",
	"WebE35":   "InternshipStart",
	"WebE35D":  "InternshipEnd",
	"WebE35T":  "InternshipScheduleChange",
	"WebE14":   "ParentalLeaveGrant",
	"WebE13":   "MotorcycleUse",
	"FastAna":  "UrgentHiring",
	"WebEA":    "AnnualLeaveNotification",
	"WebIgee":  "IGEEReport",
	"WebEpa":   "EPAReport",
	"E3":       "HireNotice",
}

// nestedNames holds the name of the public type of a nested object, per
// group. The API names the inner object of a nested object after the
// declaration type, so the public type must not carry that name.
var nestedNames = map[string]string{
	"Hire/EpikourikiSelections":         "HireBeneficiarySelection",
	"EmploymentChange/TypesMetabolon":   "ChangeTypeSelection",
	"EmploymentChange/Epikourikes":      "InsuranceSelection",
	"ConstructionWorker/Ergazomenoi":    "ConstructionEmployee",
	"ParentalLeave/Amkas":               "ParentalLeaveSocialSecurityNumber",
	"ParentalLeave/Dates":               "ParentalLeaveDate",
	"Motorcycle/Mixanes":                "Motorcycle",
	"UrgentHiring/EpikourikiSelections": "UrgentHiringBeneficiarySelection",
	"AnnualLeave/Ergazomenoi":           "AnnualLeaveEmployee",
	"IGEEReport/Details":                "IGEEReportDetail",
	"EPAReport/Details":                 "EPAReportDetail",
}

// variantSpec is a declaration type of a family: the code the API knows it
// by, the reference payload that carries it and the schema that describes it.
type variantSpec struct {
	family string
	code   string
	sample string
	xsd    string
}

// groupSpec is a family of declarations. Every group has its own type, its
// own variants and its own reference payloads.
type groupSpec struct {
	kind     string
	decl     string
	wire     string
	variant  string
	method   string
	variants []variantSpec
}

// groupSpecs holds the families of declarations the SDK models.
var groupSpecs = []groupSpec{
	{
		kind: "Hire", decl: "HireDeclaration", wire: "HirePayload",
		variant: "HireVariant", method: "SubmitHire",
		variants: []variantSpec{
			{"E3N", "WebE3N", "testProslipsiNew.json", "E3N_v2"},
			{"E3M", "WebE3M", "testMetabibashFrom.json", "E3M_v1"},
			{"E3D", "WebE3D", "testDaneismosFrom.json", "E3D_v1"},
			{"E3PD", "WebE3PD", "testProslipsiDaneismos.json", "E3PD_v2"},
		},
	},
	{
		kind: "Termination", decl: "TerminationDeclaration", wire: "TerminationPayload",
		variant: "TerminationVariant", method: "SubmitTermination",
		variants: []variantSpec{
			{"E5N", "WebE5N", "testApoxwrhshsNew.json", "E5N_v1"},
			{"E5O", "WebE5O", "testApoxwrhshOxlhsh.json", "E5O_v1"},
			{"E5AO", "WebE5AO", "testApoxwrhshsAfterOxlhsh.json", "E5AO_v1"},
			{"E5E", "WebE5E", "testApoxwrhshEthelousia.json", "E5E_v1"},
			{"E5S", "WebE5S", "testApoxwrhshsSyntaksh.json", "E5S_v1"},
			{"E5DS", "WebE5DS", "testSyntakshDekapentaetia.json", "E5DS_v1"},
			{"E5D", "WebE5D", "testApoxwrhshsThanatos.json", "E5D_v1"},
			{"E6NXP", "WebE6NXP", "testKataggeliaXwrisProidopoihsh.json", "E6NXP_v1"},
			{"E6NMP", "WebE6NMP", "testKataggeliaMeProidopoihsh.json", "E6NMP_v1"},
			{"E6SXP", "WebE6SXP", "testSyntakshXwrisProidopoihsh.json", "E6SXP_v1"},
			{"E6LT", "WebE6LT", "testLyshsDokimastikhs.json", "E6LT_v1"},
			{"E6M", "WebE6M", "testMetabibashTo.json", "E6M_v1"},
			{"E7N", "WebE7N", "testLyshSymbashsNew.json", "E7N_v1"},
		},
	},
	{
		kind: "EmploymentChange", decl: "EmploymentChangeDeclaration", wire: "EmploymentChangePayload",
		variant: "EmploymentChangeVariant", method: "SubmitEmploymentChange",
		variants: []variantSpec{
			{"MA", "WebMA", "TestMA.json", "MA_v2"},
			{"MAD", "WebMAD", "TestMAD.json", "MAD_v1"},
		},
	},
	{
		kind: "ConstructionWorker", decl: "ConstructionWorkerDeclaration", wire: "ConstructionWorkerPayload",
		variant: "ConstructionWorkerVariant", method: "SubmitConstructionWorkers",
		variants: []variantSpec{
			{"E12", "WebE12", "testE12.json", "E12_v1"},
			{"E12Ap", "WebE12Ap", "testE12Ap.json", "E12Apografiko_v1"},
		},
	},
	{
		kind: "Internship", decl: "InternshipDeclaration", wire: "InternshipPayload",
		variant: "InternshipVariant", method: "SubmitInternships",
		variants: []variantSpec{
			{"E35", "WebE35", "testE35.json", "E35_v3"},
			{"E35D", "WebE35D", "testE35D.json", "E35D_v1"},
			{"E35T", "WebE35T", "testE35T.json", "E35T_v1"},
		},
	},
	{
		kind: "ParentalLeave", decl: "ParentalLeaveDeclaration", wire: "ParentalLeavePayload",
		variant: "ParentalLeaveVariant", method: "SubmitParentalLeaves",
		variants: []variantSpec{
			{"E14", "WebE14", "testE14.json", "E14_v1"},
		},
	},
	{
		kind: "Motorcycle", decl: "MotorcycleDeclaration", wire: "MotorcyclePayload",
		variant: "MotorcycleVariant", method: "SubmitMotorcycleUse",
		variants: []variantSpec{
			{"E13", "WebE13", "testE13.json", "E13_v1"},
		},
	},
	{
		kind: "UrgentHiring", decl: "UrgentHiringDeclaration", wire: "UrgentHiringPayload",
		variant: "UrgentHiringVariant", method: "SubmitUrgentHirings",
		variants: []variantSpec{
			{"FastAna", "FastAna", "testFastAn.json", "FastAna_v1"},
		},
	},
	{
		kind: "AnnualLeave", decl: "AnnualLeaveDeclaration", wire: "AnnualLeavePayload",
		variant: "AnnualLeaveVariant", method: "SubmitAnnualLeaves",
		variants: []variantSpec{
			{"EA", "WebEA", "testEA.json", "EA_v1"},
		},
	},
	{
		kind: "IGEEReport", decl: "IGEEReportDeclaration", wire: "IGEEReportPayload",
		variant: "IGEEReportVariant", method: "SubmitIGEEReports",
		variants: []variantSpec{
			{"Igee", "WebIgee", "testIgee.json", "ApEkIgee_v1"},
		},
	},
	{
		kind: "EPAReport", decl: "EPAReportDeclaration", wire: "EPAReportPayload",
		variant: "EPAReportVariant", method: "SubmitEPAReports",
		variants: []variantSpec{
			{"Epa", "WebEpa", "testEpa.json", "ApEkEpa_v1"},
		},
	},
	{
		kind: "HireNotice", decl: "HireNoticeDeclaration", wire: "HireNoticePayload",
		variant: "HireNoticeVariant", method: "SubmitHireNotices",
		variants: []variantSpec{
			{"E3", "E3", "E3.json", "E3_v7"},
		},
	},
}

// publicTypes holds the type of a field in the public API.
var publicTypes = map[string]string{
	"string":        "string",
	"date":          "time.Time",
	"datetime":      "time.Time",
	"time":          "time.Time",
	"amount":        "float64",
	"hours":         "float64",
	"count":         "int",
	"tally":         "int",
	"number":        "int",
	"minutes":       "int",
	"flag":          "*bool",
	"base64":        "[]byte",
	"enum":          "string",
	"nested:Object": "json.RawMessage",
	"nested:List":   "json.RawMessage",
}

// wireTypes holds the type of a field on the wire.
var wireTypes = map[string]string{
	"string":        "string",
	"date":          "wire.Date",
	"datetime":      "wire.DateTime",
	"time":          "wire.Time",
	"amount":        "wire.Amount",
	"hours":         "wire.Hours",
	"count":         "wire.Count",
	"tally":         "wire.Tally",
	"number":        "wire.Plain",
	"minutes":       "wire.Minutes",
	"flag":          "wire.Bool",
	"base64":        "wire.Base64",
	"enum":          "string",
	"nested:Object": "json.RawMessage",
	"nested:List":   "json.RawMessage",
}

// optionalHelper holds the helper that converts an optional value into the
// wire encoding and the one that reads it back.
type optionalHelper struct {
	forward  string
	backward string
}

// optionalHelpers holds the helpers of the optional values, per kind.
var optionalHelpers = map[string]optionalHelper{
	"date":     {"optDate", "dateOf"},
	"datetime": {"optDateTime", "dateTimeOf"},
	"time":     {"optTime", "timeOf"},
	"amount":   {"optAmount", "amountOf"},
	"hours":    {"optHours", "hoursOf"},
	"count":    {"optCount", "countOf"},
	"tally":    {"optTally", "tallyOf"},
	"number":   {"optPlain", "plainOf"},
	"minutes":  {"optMinutes", "minutesOf"},
	"flag":     {"optBool", "boolOf"},
	"base64":   {"optBase64", "base64Of"},
	"string":   {"optString", "stringOf"},
	"enum":     {"optString", "stringOf"},
}

// optionalBody holds the body of the helpers of an optional value: the type
// in the public API, the type on the wire, the body that writes the value
// and the body that reads it.
type optionalBody struct {
	public   string
	wire     string
	toWire   string
	fromWire string
}

// optionalBodies holds the bodies of the helpers, per kind.
var optionalBodies = map[string]optionalBody{
	"date": {
		"*time.Time", "*wire.Date",
		"return &wire.Date{Time: *value}",
		"date := value.Time\n\treturn &date",
	},
	"datetime": {
		"*time.Time", "*wire.DateTime",
		"return &wire.DateTime{Time: *value}",
		"date := value.Time\n\treturn &date",
	},
	"time": {
		"*time.Time", "*wire.Time",
		"return &wire.Time{Time: *value}",
		"at := value.Time\n\treturn &at",
	},
	"amount": {
		"*float64", "*wire.Amount",
		"amount := wire.Amount(*value)\n\treturn &amount",
		"amount := float64(*value)\n\treturn &amount",
	},
	"hours": {
		"*float64", "*wire.Hours",
		"hours := wire.Hours(*value)\n\treturn &hours",
		"hours := float64(*value)\n\treturn &hours",
	},
	"count": {
		"*int", "*wire.Count",
		"count := wire.Count(*value)\n\treturn &count",
		"count := int(*value)\n\treturn &count",
	},
	"tally": {
		"*int", "*wire.Tally",
		"tally := wire.Tally(*value)\n\treturn &tally",
		"tally := int(*value)\n\treturn &tally",
	},
	"number": {
		"*int", "*wire.Plain",
		"plain := wire.Plain(*value)\n\treturn &plain",
		"plain := int(*value)\n\treturn &plain",
	},
	"minutes": {
		"*int", "*wire.Minutes",
		"minutes := wire.Minutes(*value)\n\treturn &minutes",
		"minutes := int(*value)\n\treturn &minutes",
	},
	"flag": {
		"*bool", "*wire.Bool",
		"answer := wire.FromBool(*value)\n\treturn &answer",
		"answer := *value == wire.True\n\treturn &answer",
	},
	"base64": {
		"*[]byte", "*wire.Base64",
		"document := wire.Base64FromBytes(*value)\n\treturn &document",
		"document := base64Bytes(*value)\n\treturn &document",
	},
	"string": {
		"*string", "*string",
		"text := *value\n\treturn &text",
		"text := *value\n\treturn &text",
	},
	"enum": {
		"*string", "*string",
		"text := *value\n\treturn &text",
		"text := *value\n\treturn &text",
	},
}

// convertions holds the way a value of the public API is written on the wire.
var conversions = map[string]string{
	"string":   "{v}",
	"date":     "wire.Date{Time: {v}}",
	"datetime": "wire.DateTime{Time: {v}}",
	"time":     "wire.Time{Time: {v}}",
	"amount":   "wire.Amount({v})",
	"hours":    "wire.Hours({v})",
	"count":    "wire.Count({v})",
	"tally":    "wire.Tally({v})",
	"number":   "wire.Plain({v})",
	"minutes":  "wire.Minutes({v})",
	"flag":     "wire.FromBoolPtr({v})",
	"base64":   "wire.Base64FromBytes({v})",
	"enum":     "string({v})",
	"nested":   "{v}",
}

// conversionsBack holds the way a value of the wire is read into the public
// API.
var conversionsBack = map[string]string{
	"string":   "{v}",
	"date":     "{v}.Time",
	"datetime": "{v}.Time",
	"time":     "{v}.Time",
	"amount":   "float64({v})",
	"hours":    "float64({v})",
	"count":    "int({v})",
	"tally":    "int({v})",
	"number":   "int({v})",
	"minutes":  "int({v})",
	"flag":     "boolPtr({v})",
	"base64":   "base64Bytes({v})",
	"enum":     "string({v})",
	"nested":   "{v}",
}

// The prefixes of the kinds of a field.
const (
	enumPrefix     = "enum:"
	optionalPrefix = "opt:"
)
