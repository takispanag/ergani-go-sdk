"""Generate the Go code of the declarations of the Ergani API.

One table drives the wire structs, the public models, the translation between
them and the submission methods, so the four cannot drift apart. The field
types come from the official XSD schemas and the field sets from the
reference payloads the Ministry publishes.

The generator is committed under tools/ and the code it writes is committed
too, so the build never needs it.
"""

import json
import os
import re
from collections import OrderedDict

SRC = r"C:\Users\Takis\AppData\Local\Temp\opencode\ergani-docs\json"
OUT = r"C:\Users\Takis\AppData\Local\Temp\opencode\ergani-go-sdk"

with open(r"C:\Users\Takis\AppData\Local\Temp\opencode\field-types.json", encoding="utf-8") as fh:
    SCHEMA_TYPES = json.load(fh)
with open(r"C:\Users\Takis\AppData\Local\Temp\opencode\field-docs.json", encoding="utf-8") as fh:
    DOCS = json.load(fh)

# The name of the field in the public API, and its type, for the fields that
# are known. The type is the one the hand written table gives; every other
# field is inferred from the schema of its declaration.
KNOWN = {
    "f_aa_pararthmatos": ("string", "BranchNumber"),
    "f_rel_protocol": ("string", "RelatedProtocol"),
    "f_rel_date": ("date", "RelatedDate"),
    "f_ypiresia_sepe": ("string", "SEPEServiceCode"),
    "f_ypiresia_oaed": ("string", "OAEEDServiceCode"),
    "f_kad_pararthmatos": ("string", "BranchActivityCode"),
    "f_kallikratis_pararthmatos": ("string", "MunicipalCode"),
    "f_eponymo": ("string", "LastName"),
    "f_onoma": ("string", "FirstName"),
    "f_onoma_patros": ("string", "FatherName"),
    "f_onoma_mitros": ("string", "MotherName"),
    "f_birthdate": ("date", "BirthDate"),
    "f_sex": ("enum:Sex", "Sex"),
    "f_yphkoothta": ("string", "NationalityCode"),
    "f_typos_taytothtas": ("string", "IdentityDocumentType"),
    "f_ar_taytothtas": ("string", "IdentityDocumentNumber"),
    "f_ekdousa_arxh": ("string", "IdentityDocumentIssuer"),
    "f_date_ekdosis": ("date", "IdentityDocumentIssued"),
    "f_date_ekdosis_lixi": ("date", "IdentityDocumentExpires"),
    "f_res_permit_inst": ("flag", "HasResidencePermit"),
    "f_res_permit_inst_type": ("string", "ResidencePermitType"),
    "f_res_permit_inst_ar": ("string", "ResidencePermitNumber"),
    "f_res_permit_inst_lixi": ("date", "ResidencePermitExpires"),
    "f_res_permit_ap": ("flag", "HasWorkPermit"),
    "f_res_permit_ap_type": ("string", "WorkPermitType"),
    "f_res_permit_ap_ar": ("string", "WorkPermitNumber"),
    "f_res_permit_ap_lixi": ("date", "WorkPermitExpires"),
    "f_res_permit_visa": ("flag", "HasVisa"),
    "f_res_permit_visa_ar": ("string", "VisaNumber"),
    "f_res_permit_visa_from": ("date", "VisaFrom"),
    "f_res_permit_visa_to": ("date", "VisaTo"),
    "f_marital_status": ("enum:MaritalStatus", "MaritalStatus"),
    "f_arithmos_teknon": ("count", "Children"),
    "f_afm": ("string", "TaxID"),
    "f_doy": ("string", "TaxOffice"),
    "f_amika": ("string", "SocialSecurityRegistryNumber"),
    "f_amka": ("string", "SocialSecurityNumber"),
    "f_code_anergias": ("string", "UnemploymentCode"),
    "f_ar_vivliou_anilikou": ("string", "AnnualIncome"),
    "f_epipedo_morfosis": ("string", "EducationLevel"),
    "f_xaraktirismos": ("enum:WorkerCategory", "WorkerCategory"),
    "f_sxeshapasxolisis": ("enum:RelationshipType", "RelationshipType"),
    "f_orismenou_apo": ("date", "FixedTermFrom"),
    "f_orismenou_ews": ("date", "FixedTermTo"),
    "f_proslipsidate": ("date", "HireDate"),
    "f_proslipsitime": ("time", "HireTime"),
    "f_apoxwrisidate": ("date", "LeaveDate"),
    "f_apoxwrisitime": ("time", "LeaveTime"),
    "f_eidikothta": ("string", "ProfessionCode"),
    "f_eidikothta_anal": ("string", "ProfessionDescription"),
    "f_proipiresia": ("string", "InsuranceFund"),
    "f_apodoxes": ("amount", "MonthlyWages"),
    "f_hour_apodoxes": ("amount", "HourlyWages"),
    "f_kathestosapasxolisis": ("enum:EmploymentStatus", "EmploymentStatus"),
    "f_week_hours": ("hours", "WeeklyHours"),
    "f_full_employment_hours": ("hours", "FullEmploymentHours"),
    "f_week_days": ("enum:WeekPattern", "WeeklyPattern"),
    "f_working_card": ("flag", "UsesWorkCard"),
    "f_working_time_digital_organization": ("flag", "UsesDigitalWorkingTime"),
    "f_dialeimma_minutes": ("minutes", "BreakMinutes"),
    "f_dialeimma_entos_wrariou": ("flag", "BreakWithinSchedule"),
    "f_euelikto_wrario_minutes": ("minutes", "FlexibleArrivalMinutes"),
    "f_responsible_position": ("string", "ResponsiblePosition"),
    "f_kyria_asfalisi": ("string", "PrimaryInsuranceFund"),
    "f_prosthetes_asfalistikes": ("string", "SupplementaryInsuranceFund"),
    "f_ipoxreotiki_katartisi": ("string", "RehabilitationRecord"),
    "f_efarmoste_sillogiki_simbasi": ("flag", "HasCollectiveAgreement"),
    "f_efarmoste_sillogiki_simbasi_comments": ("string", "CollectiveAgreementNotes"),
    "f_efarmostea_sillogiki_simbasi": ("flag", "HasSignedCollectiveAgreement"),
    "f_efarmostea_sillogiki_simbasi_comments": ("string", "SignedCollectiveAgreementNotes"),
    "f_mh_provlepsimo_programma": ("flag", "HasUnlistedProgram"),
    "f_topos_ergasias": ("flag", "HasDeclaredPlaceOfWork"),
    "f_topos_ergasias_comment": ("string", "PlaceOfWorkNotes"),
    "f_paraggelia_hmeres_hours": ("string", "NoticeDaysAndHours"),
    "f_paraggelia_min_notification": ("string", "MinimumNoticePeriod"),
    "f_paraggelia_notes": ("string", "NoticeNotes"),
    "f_xronos_katavolis_apodoxon": ("string", "WagesPaymentFrequency"),
    "f_comments": ("string", "Comments"),
    "f_file": ("base64", "File"),
    "f_foreign_file": ("base64", "ForeignWorkerFile"),
    "f_young_file": ("base64", "MinorWorkerFile"),
    "f_file_symbash": ("base64", "IndividualAgreementFile"),
    "f_basics_acceptance": ("enum:EssentialTermsAcceptance", "EssentialTermsAcceptance"),
    "f_atomikh_symbash": ("enum:IndividualAgreementStatus", "IndividualAgreementStatus"),
    "f_special_case": ("string", "SpecialCase"),
    "f_trial_period": ("flag", "HasTrialPeriod"),
    "f_trial_date_to": ("date", "TrialPeriodEnds"),
    "f_topothetisioaed": ("flag", "HasPriorProgramPlacement"),
    "f_programaoaed": ("string", "ProgramPlacement"),
    "f_replaceprograma": ("flag", "ReplacesBeneficiary"),
    "f_replaceprograma_afm": ("string", "ReplacedBeneficiaryTaxID"),
    "f_replaceprograma_amka": ("string", "ReplacedBeneficiarySocialSecurityNumber"),
    "f_transfer_company_afm": ("string", "TransferCompanyTaxID"),
    "f_transfer_company_eponimia": ("string", "TransferCompanyName"),
    "f_date_metabibashs": ("date", "TransferDate"),
    "f_borrow_company_afm": ("string", "BorrowingCompanyTaxID"),
    "f_borrow_company_eponimia": ("string", "BorrowingCompanyName"),
    "f_borrow_date_from": ("date", "LoanFrom"),
    "f_borrow_date_to": ("date", "LoanTo"),
    "f_borrow_type": ("string", "LoanType"),
    "f_ergazom_borrow_agreement": ("string", "BorrowedStaffAgreement"),
    "f_kataboli_apodoxon": ("string", "WagesPaymentMethod"),
    "f_apolysisdate": ("date", "TerminationDate"),
    "f_lixisymbashdate": ("date", "ContractEndDate"),
    "f_logosperatosis": ("string", "TerminationReason"),
    "f_logosperatosiscomments": ("string", "TerminationReasonNotes"),
    "f_oros": ("string", "Month"),
    "f_omadiki": ("flag", "HasCollectiveAction"),
    "f_omadikidate": ("date", "CollectiveActionDate"),
    "f_omadikiarithmos": ("string", "CollectiveActionNumber"),
    "f_minesproidopoihsh": ("string", "NotificationMonths"),
    "f_proidopoihshdate": ("date", "NoticeDate"),
    "f_koinopoihshdate": ("date", "DismissalDate"),
    "f_posoapozimiosis": ("amount", "DismissalCompensation"),
    "f_tropoi_oxlhshs": ("string", "ResignationMethods"),
    "f_oxlhsh_date_ypovolis": ("date", "NotificationServiceDate"),
    "f_oxlhsh_protocol": ("string", "NotificationServiceProtocol"),
}

# The abbreviations the public API spells in capitals.
INITIALISMS = {
    "afm": "AFM", "amka": "AMKA", "amika": "AMIKA", "sepe": "SEPE",
    "oaed": "OAED", "kad": "KAD", "id": "ID", "url": "URL", "pdf": "PDF",
    "vat": "VAT", "iban": "IBAN", "covid": "COVID", "eae": "EAE",
}

# The name of the constant of a variant, when the code alone would not say
# what the declaration is.
SEMANTIC = {
    "WebE3N": "HireNew",
    "WebE3M": "HireTransfer",
    "WebE3D": "HireBorrowed",
    "WebE3PD": "HireForLoanOut",
    "WebE5N": "TerminationResignation",
    "WebE5O": "TerminationNotice",
    "WebE5AO": "TerminationResignationAfterNotice",
    "WebE5E": "TerminationVoluntaryExit",
    "WebE5S": "TerminationRetirement",
    "WebE5DS": "TerminationFifteenYears",
    "WebE5D": "TerminationDeath",
    "WebE6NXP": "TerminationDismissalWithoutNotice",
    "WebE6NMP": "TerminationDismissalWithNotice",
    "WebE6SXP": "TerminationRetirementByDismissal",
    "WebE6LT": "TerminationProbation",
    "WebE6M": "TerminationTransfer",
    "WebE7N": "TerminationFixedTerm",
    "WebMA": "ChangeEmployment",
    "WebMAD": "ChangeBorrowedEmployment",
    "WebE12": "ConstructionWorker",
    "WebE12Ap": "ConstructionWorkerCensus",
    "WebE35": "InternshipStart",
    "WebE35D": "InternshipEnd",
    "WebE35T": "InternshipScheduleChange",
    "WebE14": "ParentalLeaveGrant",
    "WebE13": "MotorcycleUse",
    "FastAna": "UrgentHiring",
    "WebEA": "AnnualLeaveNotification",
    "WebIgee": "IGEEReport",
    "WebEpa": "EPAReport",
    "E3": "HireNotice",
}

# The name of the public type of a nested object, per group. The API names
# the inner object of a nested object after the declaration type, so the
# public type must not carry that name.
NESTED_NAMES = {
    ("Hire", "EpikourikiSelections"): "HireBeneficiarySelection",
    ("EmploymentChange", "TypesMetabolon"): "ChangeTypeSelection",
    ("EmploymentChange", "Epikourikes"): "InsuranceSelection",
    ("ConstructionWorker", "Ergazomenoi"): "ConstructionEmployee",
    ("ParentalLeave", "Amkas"): "ParentalLeaveSocialSecurityNumber",
    ("ParentalLeave", "Dates"): "ParentalLeaveDate",
    ("Motorcycle", "Mixanes"): "Motorcycle",
    ("UrgentHiring", "EpikourikiSelections"): "UrgentHiringBeneficiarySelection",
    ("AnnualLeave", "Ergazomenoi"): "AnnualLeaveEmployee",
    ("IGEEReport", "Details"): "IGEEReportDetail",
    ("EPAReport", "Details"): "EPAReportDetail",
}

# The declaration groups. Every group has its own type, its own variants and
# its own reference payloads.
GROUPS = [
    {
        "kind": "Hire",
        "struct": "HireDeclaration",
        "wire": "HirePayload",
        "variant": "HireVariant",
        "method": "SubmitHire",
        "variants": [
            ("E3N", "WebE3N", "testProslipsiNew.json", "E3N_v2"),
            ("E3M", "WebE3M", "testMetabibashFrom.json", "E3M_v1"),
            ("E3D", "WebE3D", "testDaneismosFrom.json", "E3D_v1"),
            ("E3PD", "WebE3PD", "testProslipsiDaneismos.json", "E3PD_v2"),
        ],
    },
    {
        "kind": "Termination",
        "struct": "TerminationDeclaration",
        "wire": "TerminationPayload",
        "variant": "TerminationVariant",
        "method": "SubmitTermination",
        "variants": [
            ("E5N", "WebE5N", "testApoxwrhshsNew.json", "E5N_v1"),
            ("E5O", "WebE5O", "testApoxwrhshOxlhsh.json", "E5O_v1"),
            ("E5AO", "WebE5AO", "testApoxwrhshsAfterOxlhsh.json", "E5AO_v1"),
            ("E5E", "WebE5E", "testApoxwrhshEthelousia.json", "E5E_v1"),
            ("E5S", "WebE5S", "testApoxwrhshsSyntaksh.json", "E5S_v1"),
            ("E5DS", "WebE5DS", "testSyntakshDekapentaetia.json", "E5DS_v1"),
            ("E5D", "WebE5D", "testApoxwrhshsThanatos.json", "E5D_v1"),
            ("E6NXP", "WebE6NXP", "testKataggeliaXwrisProidopoihsh.json", "E6NXP_v1"),
            ("E6NMP", "WebE6NMP", "testKataggeliaMeProidopoihsh.json", "E6NMP_v1"),
            ("E6SXP", "WebE6SXP", "testSyntakshXwrisProidopoihsh.json", "E6SXP_v1"),
            ("E6LT", "WebE6LT", "testLyshsDokimastikhs.json", "E6LT_v1"),
            ("E6M", "WebE6M", "testMetabibashTo.json", "E6M_v1"),
            ("E7N", "WebE7N", "testLyshSymbashsNew.json", "E7N_v1"),
        ],
    },
    {
        "kind": "EmploymentChange",
        "struct": "EmploymentChangeDeclaration",
        "wire": "EmploymentChangePayload",
        "variant": "EmploymentChangeVariant",
        "method": "SubmitEmploymentChange",
        "variants": [
            ("MA", "WebMA", "TestMA.json", "MA_v2"),
            ("MAD", "WebMAD", "TestMAD.json", "MAD_v1"),
        ],
    },
    {
        "kind": "ConstructionWorker",
        "struct": "ConstructionWorkerDeclaration",
        "wire": "ConstructionWorkerPayload",
        "variant": "ConstructionWorkerVariant",
        "method": "SubmitConstructionWorkers",
        "variants": [
            ("E12", "WebE12", "testE12.json", "E12_v1"),
            ("E12Ap", "WebE12Ap", "testE12Ap.json", "E12Apografiko_v1"),
        ],
    },
    {
        "kind": "Internship",
        "struct": "InternshipDeclaration",
        "wire": "InternshipPayload",
        "variant": "InternshipVariant",
        "method": "SubmitInternships",
        "variants": [
            ("E35", "WebE35", "testE35.json", "E35_v3"),
            ("E35D", "WebE35D", "testE35D.json", "E35D_v1"),
            ("E35T", "WebE35T", "testE35T.json", "E35T_v1"),
        ],
    },
    {
        "kind": "ParentalLeave",
        "struct": "ParentalLeaveDeclaration",
        "wire": "ParentalLeavePayload",
        "variant": "ParentalLeaveVariant",
        "method": "SubmitParentalLeaves",
        "variants": [
            ("E14", "WebE14", "testE14.json", "E14_v1"),
        ],
    },
    {
        "kind": "Motorcycle",
        "struct": "MotorcycleDeclaration",
        "wire": "MotorcyclePayload",
        "variant": "MotorcycleVariant",
        "method": "SubmitMotorcycleUse",
        "variants": [
            ("E13", "WebE13", "testE13.json", "E13_v1"),
        ],
    },
    {
        "kind": "UrgentHiring",
        "struct": "UrgentHiringDeclaration",
        "wire": "UrgentHiringPayload",
        "variant": "UrgentHiringVariant",
        "method": "SubmitUrgentHirings",
        "variants": [
            ("FastAna", "FastAna", "testFastAn.json", "FastAna_v1"),
        ],
    },
    {
        "kind": "AnnualLeave",
        "struct": "AnnualLeaveDeclaration",
        "wire": "AnnualLeavePayload",
        "variant": "AnnualLeaveVariant",
        "method": "SubmitAnnualLeaves",
        "variants": [
            ("EA", "WebEA", "testEA.json", "EA_v1"),
        ],
    },
    {
        "kind": "IGEEReport",
        "struct": "IGEEReportDeclaration",
        "wire": "IGEEReportPayload",
        "variant": "IGEEReportVariant",
        "method": "SubmitIGEEReports",
        "variants": [
            ("Igee", "WebIgee", "testIgee.json", "ApEkIgee_v1"),
        ],
    },
    {
        "kind": "EPAReport",
        "struct": "EPAReportDeclaration",
        "wire": "EPAReportPayload",
        "variant": "EPAReportVariant",
        "method": "SubmitEPAReports",
        "variants": [
            ("Epa", "WebEpa", "testEpa.json", "ApEkEpa_v1"),
        ],
    },
    {
        "kind": "HireNotice",
        "struct": "HireNoticeDeclaration",
        "wire": "HireNoticePayload",
        "variant": "HireNoticeVariant",
        "method": "SubmitHireNotices",
        "variants": [
            ("E3", "E3", "E3.json", "E3_v7"),
        ],
    },
]

PUBLIC_TYPE = {
    "string": "string",
    "date": "time.Time",
    "datetime": "time.Time",
    "time": "time.Time",
    "amount": "float64",
    "hours": "float64",
    "count": "int",
    "tally": "int",
    "number": "int",
    "minutes": "int",
    "flag": "*bool",
    "base64": "[]byte",
    "enum": "string",
    "nested:Object": "json.RawMessage",
    "nested:List": "json.RawMessage",
}

WIRE_TYPE = {
    "string": "string",
    "date": "wire.Date",
    "datetime": "wire.DateTime",
    "time": "wire.Time",
    "amount": "wire.Amount",
    "hours": "wire.Hours",
    "count": "wire.Count",
    "tally": "wire.Tally",
    "number": "wire.Plain",
    "minutes": "wire.Minutes",
    "flag": "wire.Bool",
    "base64": "wire.Base64",
    "enum": "string",
    "nested:Object": "json.RawMessage",
    "nested:List": "json.RawMessage",
}

# The helper that converts an optional value, and the one that reads it back.
OPT = {
    "date": ("optDate", "dateOf"),
    "datetime": ("optDateTime", "dateTimeOf"),
    "time": ("optTime", "timeOf"),
    "amount": ("optAmount", "amountOf"),
    "hours": ("optHours", "hoursOf"),
    "count": ("optCount", "countOf"),
    "tally": ("optTally", "tallyOf"),
    "number": ("optPlain", "plainOf"),
    "minutes": ("optMinutes", "minutesOf"),
    "flag": ("optBool", "boolOf"),
    "base64": ("optBase64", "base64Of"),
    "string": ("optString", "stringOf"),
    "enum": ("optString", "stringOf"),
}

# The bodies of the helpers, per kind.
OPT_BODIES = {
    "date": ("*time.Time", "*wire.Date", "return &wire.Date{Time: *value}", "date := value.Time\n\treturn &date"),
    "datetime": ("*time.Time", "*wire.DateTime", "return &wire.DateTime{Time: *value}", "date := value.Time\n\treturn &date"),
    "time": ("*time.Time", "*wire.Time", "return &wire.Time{Time: *value}", "at := value.Time\n\treturn &at"),
    "amount": ("*float64", "*wire.Amount", "amount := wire.Amount(*value)\n\treturn &amount", "amount := float64(*value)\n\treturn &amount"),
    "hours": ("*float64", "*wire.Hours", "hours := wire.Hours(*value)\n\treturn &hours", "hours := float64(*value)\n\treturn &hours"),
    "count": ("*int", "*wire.Count", "count := wire.Count(*value)\n\treturn &count", "count := int(*value)\n\treturn &count"),
    "tally": ("*int", "*wire.Tally", "tally := wire.Tally(*value)\n\treturn &tally", "tally := int(*value)\n\treturn &tally"),
    "number": ("*int", "*wire.Plain", "plain := wire.Plain(*value)\n\treturn &plain", "plain := int(*value)\n\treturn &plain"),
    "minutes": ("*int", "*wire.Minutes", "minutes := wire.Minutes(*value)\n\treturn &minutes", "minutes := int(*value)\n\treturn &minutes"),
    "flag": ("*bool", "*wire.Bool", "answer := wire.FromBool(*value)\n\treturn &answer", "answer := *value == wire.True\n\treturn &answer"),
    "base64": ("*[]byte", "*wire.Base64", "document := wire.Base64FromBytes(*value)\n\treturn &document", "document := base64Bytes(*value)\n\treturn &document"),
    "string": ("*string", "*string", "text := *value\n\treturn &text", "text := *value\n\treturn &text"),
    "enum": ("*string", "*string", "text := *value\n\treturn &text", "text := *value\n\treturn &text"),
}

# The kinds of the optional values the models use.
used_opt = set()

CONVERT = {
    "string": "{v}",
    "date": "wire.Date{Time: {v}}",
    "datetime": "wire.DateTime{Time: {v}}",
    "time": "wire.Time{Time: {v}}",
    "amount": "wire.Amount({v})",
    "hours": "wire.Hours({v})",
    "count": "wire.Count({v})",
    "tally": "wire.Tally({v})",
    "number": "wire.Plain({v})",
    "minutes": "wire.Minutes({v})",
    "flag": "wire.FromBoolPtr({v})",
    "base64": "wire.Base64FromBytes({v})",
    "enum": "string({v})",
    "nested": "{v}",
}

CONVERT_BACK = {
    "string": "{v}",
    "date": "{v}.Time",
    "datetime": "{v}.Time",
    "time": "{v}.Time",
    "amount": "float64({v})",
    "hours": "float64({v})",
    "count": "int({v})",
    "tally": "int({v})",
    "number": "int({v})",
    "minutes": "int({v})",
    "flag": "boolPtr({v})",
    "base64": "base64Bytes({v})",
    "enum": "string({v})",
    "nested": "{v}",
}

ENUM_PREFIX = "enum:"
OPT_PREFIX = "opt:"


def public_type(kind):
    if kind.startswith(OPT_PREFIX):
        base = kind[len(OPT_PREFIX):]
        # A boolean is already a pointer, so it carries the three states on
        # its own and an optional one is the same.
        if base == "flag":
            return "*bool"
        return "*" + public_type(base)
    if kind.startswith(ENUM_PREFIX):
        return kind[len(ENUM_PREFIX):]
    if kind.startswith("nested:"):
        return "json.RawMessage"
    return PUBLIC_TYPE[kind]


def wire_type(kind):
    if kind.startswith(OPT_PREFIX):
        base = kind[len(OPT_PREFIX):]
        if base == "flag":
            return "*wire.Bool"
        return "*" + wire_type(base)
    if kind.startswith(ENUM_PREFIX):
        return "string"
    if kind.startswith("nested:"):
        return "json.RawMessage"
    return WIRE_TYPE[kind]


def wire_type_local(kind):
    if kind.startswith(OPT_PREFIX):
        base = kind[len(OPT_PREFIX):]
        if base == "flag":
            return "*Bool"
        return "*" + wire_type_local(base)
    if kind.startswith(ENUM_PREFIX):
        return "string"
    if kind.startswith("nested:"):
        return "json.RawMessage"
    return WIRE_TYPE[kind].replace("wire.", "")


def convert(kind):
    if kind.startswith(OPT_PREFIX):
        base = kind[len(OPT_PREFIX):]
        used_opt.add(base)
        if base == "flag":
            return "optBoolPtr({v})"
        if base.startswith(ENUM_PREFIX):
            return f"optString((*string)({{v}}))"
        return f"{OPT[base][0]}({{v}})"
    if kind.startswith(ENUM_PREFIX):
        return "string({v})"
    return CONVERT[kind]


def convert_back(kind):
    if kind.startswith(OPT_PREFIX):
        base = kind[len(OPT_PREFIX):]
        used_opt.add(base)
        if base == "flag":
            return "boolPtrOf({v})"
        if base.startswith(ENUM_PREFIX):
            return f"optString((*string)({{v}}))"
        return f"{OPT[base][1]}({{v}})"
    if kind.startswith(ENUM_PREFIX):
        return f"{kind[len(ENUM_PREFIX):]}({{v}})"
    return CONVERT_BACK[kind]


def go_name(field):
    """The name of a field in the public API, when it is not known."""
    if field.startswith("f_"):
        field = field[2:]
    parts = field.split("_")
    name = ""
    for part in parts:
        if not part:
            continue
        name += INITIALISMS.get(part, part[:1].upper() + part[1:])
    return name


def field_kind(field, xsd, value=None, optional=False):
    """The type of a field, and its name in the public API."""
    base = None
    name = None

    if isinstance(value, dict) and value.get("nested"):
        base = "nested:List" if isinstance(value.get("inner"), list) else "nested:Object"
        name = go_name(field)
    elif field in KNOWN:
        base, name = KNOWN[field]
    else:
        base = SCHEMA_TYPES.get(xsd, {}).get(field, "string")
        name = go_name(field)

    # A schema that allows one to three digits does not say how many the
    # API writes, so the reference payload decides.
    if base in ("count", "tally") and isinstance(value, dict):
        text = value.get("value")
        if isinstance(text, str) and text.isdigit():
            base = "tally" if len(text) == 3 else "count"

    if optional:
        return ("opt:" + base, name)

    return (base, name)


def inner_fields(entry):
    """The fields of the inner object of a nested field.

    The entries of a list do not always carry the same fields, so the inner
    object is the union of them.
    """
    inner = entry.get("inner")
    if isinstance(inner, list):
        union = {}
        for element in inner:
            if not isinstance(element, dict):
                continue
            for key, item in element.items():
                if key in union:
                    continue
                if isinstance(item, dict) and item:
                    nested_key = list(item)[0]
                    union[key] = {
                        "nested": True,
                        "inner_key": nested_key,
                        "inner": item[nested_key],
                    }
                else:
                    union[key] = {"nested": False, "value": item}
        return union
    if not isinstance(inner, dict):
        return {}

    fields = {}
    for key, item in inner.items():
        if isinstance(item, dict) and item:
            nested_key = list(item)[0]
            fields[key] = {
                "nested": True,
                "inner_key": nested_key,
                "inner": item[nested_key],
            }
        else:
            fields[key] = {"nested": False, "value": item}
    return fields


def wire_nested_name(kind, field):
    """The name of the wire type of a nested object."""
    return kind + go_name(field)


def nested_name(kind, field):
    """The name of the public type of a nested object."""
    return NESTED_NAMES.get((kind, field), go_name(field))


def nested_entry(entry, field):
    """The nested object of a field, if it has one."""
    for variant in entry["variants"]:
        info = variant["fields"].get(field)
        if info and info.get("nested"):
            return info
    return None


def is_list(info):
    return isinstance(info.get("inner"), list)


def nested_variants(entry, field):
    """The inner fields of a nested field, per variant.

    The variants of a family do not always carry the same inner fields, so
    the inner object is the union of them and every variant sends the ones it
    holds.
    """
    result = {}
    for variant in entry["variants"]:
        info = variant["fields"].get(field)
        if not info or not info.get("nested"):
            continue
        if info.get("entries"):
            # The entries of a list do not always carry the same fields, so
            # the inner object is the union of them.
            union = {}
            for element in info["entries"]:
                for name, item in element.items():
                    if name not in union:
                        union[name] = {"nested": False, "value": item}
            result[variant["code"]] = union
            continue
        result[variant["code"]] = inner_fields(info)
    return result


def nested_union(per_variant):
    """The union of the inner fields of a nested object."""
    union = []
    for fields in per_variant.values():
        for name in fields:
            if name not in union:
                union.append(name)
    return union


def nested_union_fields(entry, field):
    """The union of the inner fields of a nested object, with their values."""
    union = []
    seen = set()
    for variant in entry["variants"]:
        info = variant["fields"].get(field)
        if not info or not info.get("nested"):
            continue
        for name, item in inner_fields(info).items():
            if name in seen:
                continue
            seen.add(name)
            union.append((name, item))
    return union


def inner_presence(entry, field):
    """The fields of a nested object that every one of them carries.

    A field that only some of the variants, or some of the entries of a list,
    carry is optional: the API leaves it out, and the SDK has to leave it out
    too.
    """
    sets = []
    for variant in entry["variants"]:
        info = variant["fields"].get(field)
        if not info or not info.get("nested"):
            continue
        if info.get("entries"):
            for element in info["entries"]:
                sets.append(set(element))
        else:
            sets.append(set(inner_fields(info)))

    if not sets:
        return set()

    required = set(sets[0])
    for names in sets[1:]:
        required &= names

    return required


def leaves(node, path, out):
    if isinstance(node, dict):
        for key, value in node.items():
            leaves(value, f"{path}.{key}" if path else key, out)
    elif isinstance(node, list):
        if node:
            leaves(node[0], f"{path}[]", out)
    else:
        out[path] = node


def fields_of(filename):
    """The fields of a reference payload, and the keys of its envelope.

    A field whose value is an object is a nested object. The API names its
    inner object after the declaration type, so the inner key is kept with the
    field and the public type does not carry it.
    """
    with open(os.path.join(SRC, filename), encoding="utf-8-sig") as fh:
        data = json.load(fh)

    root = list(data)[0]
    inner = data[root]
    inner_key = list(inner)[0] if isinstance(inner, dict) else ""
    value = inner[inner_key] if inner_key else inner

    if isinstance(value, list):
        value = value[0] if value else {}
    if not isinstance(value, dict):
        value = {}

    fields = {}
    for key, item in value.items():
        if isinstance(item, dict) and item:
            nested_key = list(item)[0]
            inner = item[nested_key]
            entries = []
            if isinstance(inner, list):
                for element in inner:
                    if isinstance(element, dict):
                        entries.append(element)
            fields[key] = {
                "nested": True,
                "inner_key": nested_key,
                "inner": inner,
                "entries": entries,
            }
        else:
            fields[key] = {"nested": False, "value": item}

    return fields, root, inner_key


# The fields of every group, and the variants that carry them.
groups = []
for group in GROUPS:
    variants = []
    for family, code, sample, xsd in group["variants"]:
        fields, root, inner = fields_of(sample)
        variants.append({
            "family": family,
            "code": code,
            "sample": sample,
            "xsd": xsd,
            "fields": fields,
            "root": root,
            "inner": inner,
        })

    union = []
    owner = {}
    for variant in variants:
        for field in variant["fields"]:
            if field not in union:
                union.append(field)
            owner.setdefault(field, set()).add(variant["family"])

    common = set.intersection(*(set(v["fields"]) for v in variants))
    shared = [f for f in union if f in common]
    optional = [f for f in union if f not in common]

    groups.append({
        "group": group,
        "variants": variants,
        "union": union,
        "owner": owner,
        "shared": shared,
        "optional": optional,
    })

# ---------------------------------------------------------------------------
# The wire structs.
# ---------------------------------------------------------------------------

wire = []
wire.append("// Code generated by tools/generate_declarations.py. DO NOT EDIT.")
wire.append("")
wire.append("package wire")
wire.append("")

for entry in groups:
    group = entry["group"]
    group_xsd = entry["variants"][0]["xsd"]
    wire.append(f"// {group['wire']} is the flat set of the fields of a {group['kind'].lower()} declaration.")
    wire.append(f"type {group['wire']} struct {{")
    for field in entry["shared"] + entry["optional"]:
        info = entry["variants"][0]["fields"].get(field)
        fieldkind, name = field_kind(field, entry["variants"][0]["xsd"], info)
        if fieldkind.startswith("nested:"):
            sub = wire_nested_name(group["kind"], field)
            wire.append(f'\t{name} {sub}Payload `json:"{field}"`')
            continue
        wire.append(f'\t{name} {wire_type_local(fieldkind)} `json:"{field}"`')
    wire.append("}")
    wire.append("")

# The types of the nested objects. The API names the inner object after the
# declaration type, so the wire keeps a map keyed by that name.
seen_wire = {}
for entry in groups:
    kind = entry["group"]["kind"]
    for field in entry["shared"] + entry["optional"]:
        info = nested_entry(entry, field)
        if not info:
            continue
        sub = wire_nested_name(kind, field)
        if sub in seen_wire:
            continue
        seen_wire[sub] = True

        inner = inner_fields(info)
        inner_name = sub + "Entry" if is_list(info) else sub

        wire.append(f"// {sub}Payload is a nested object of a {kind.lower()} declaration.")
        if is_list(info):
            wire.append(f"type {sub}Payload map[string][]{inner_name}")
        else:
            wire.append(f"type {sub}Payload map[string]{inner_name}")
        wire.append("")
        wire.append(f"// {inner_name} is the inner object of a nested object of a {kind.lower()} declaration.")
        wire.append(f"type {inner_name} struct {{")
        required = inner_presence(entry, field)
        for inner_field, inner_info in nested_union_fields(entry, field):
            innerkind, inner_field_name = field_kind(
                inner_field, entry["variants"][0]["xsd"], inner_info,
                optional=inner_field not in required)
            omit = ",omitempty" if innerkind.startswith(OPT_PREFIX) else ""
            wire.append(f'\t{inner_field_name} {wire_type_local(innerkind)} `json:"{inner_field}{omit}"`')
        wire.append("}")
        wire.append("")

# ---------------------------------------------------------------------------
# The public models.
# ---------------------------------------------------------------------------

# The public models, and the types of the nested objects.

public = []
public.append("// Code generated by tools/generate_declarations.py. DO NOT EDIT.")
public.append("")
public.append("package ergani")
public.append("")
public.append('import "time"')
public.append("")

for entry in groups:
    group = entry["group"]
    kind = group["kind"]
    public.append(f"// {group['struct']} is the declaration of the {kind.lower()} of an employer.")
    public.append(f"type {group['struct']} struct {{")
    for field in entry["shared"] + entry["optional"]:
        info = entry["variants"][0]["fields"].get(field)
        fieldkind, name = field_kind(field, entry["variants"][0]["xsd"], info)
        doc = DOCS.get(field, "")
        if doc:
            public.append(f"\t// {name} — {doc}")
        if fieldkind.startswith("nested:"):
            sub = nested_name(kind, field)
            pointer = "[]" if fieldkind == "nested:List" else "*"
            public.append(f"\t{name} {pointer}{sub}")
            continue
        public.append(f"\t{name} {public_type(fieldkind)}")
    public.append("}")
    public.append("")

# The types of the nested objects.
seen_nested = {}
for entry in groups:
    kind = entry["group"]["kind"]
    for field in entry["shared"] + entry["optional"]:
        info = nested_entry(entry, field)
        if not info:
            continue
        sub = nested_name(kind, field)
        if sub in seen_nested:
            continue
        seen_nested[sub] = True

        public.append(f"// {sub} is a part of a {kind.lower()} declaration.")
        public.append(f"type {sub} struct {{")
        required = inner_presence(entry, field)
        for inner_field, inner_info in nested_union_fields(entry, field):
            innerkind, inner_name = field_kind(
                inner_field, entry["variants"][0]["xsd"], inner_info,
                optional=inner_field not in required)
            inner_doc = DOCS.get(inner_field, "")
            if inner_doc:
                public.append(f"\t// {inner_name} — {inner_doc}")
            public.append(f"\t{inner_name} {public_type(innerkind)}")
        public.append("}")
        public.append("")

# ---------------------------------------------------------------------------
# The translation and the submission.
# ---------------------------------------------------------------------------

mapping = []

for entry in groups:
    group = entry["group"]
    kind = group["kind"]
    struct = group["struct"]
    wire_name = group["wire"]
    variant = group["variant"]
    method = group["method"]
    group_xsd = entry["variants"][0]["xsd"]
    lower = kind[0].lower() + kind[1:]
    payload_of = kind[0].lower() + kind[1:] + "PayloadOf"
    declaration_of = kind[0].lower() + kind[1:] + "DeclarationOf"

    mapping.append(f"// {payload_of} turns a {struct} into the payload of the API.")
    mapping.append(f"func {payload_of}(d {struct}) wire.{wire_name} {{")
    mapping.append(f"\tp := wire.{wire_name}{{}}")
    for field in entry["shared"] + entry["optional"]:
        info = entry["variants"][0]["fields"].get(field)
        fieldkind, name = field_kind(field, entry["variants"][0]["xsd"], info)
        if fieldkind.startswith("nested:"):
            # The inner object of a nested object is named after the
            # declaration type, so it is built where the variant is known.
            continue
        mapping.append(f"\tp.{name} = " + convert(fieldkind).replace("{v}", f"d.{name}"))
    mapping.append("\treturn p")
    mapping.append("}")
    mapping.append("")

    mapping.append(f"// {declaration_of} turns a payload of the API into a {struct}.")
    mapping.append(f"func {declaration_of}(p wire.{wire_name}) {struct} {{")
    mapping.append(f"\tvar d {struct}")
    for field in entry["shared"] + entry["optional"]:
        info = entry["variants"][0]["fields"].get(field)
        fieldkind, name = field_kind(field, entry["variants"][0]["xsd"], info)
        if fieldkind.startswith("nested:"):
            sub = nested_name(kind, field)
            helper = lower + sub
            mapping.append(f"\td.{name} = {helper}Of(p.{name})")
            continue
        mapping.append(f"\td.{name} = " + convert_back(fieldkind).replace("{v}", f"p.{name}"))
    mapping.append("\treturn d")
    mapping.append("}")
    mapping.append("")

    # The helpers of the nested objects. The API names the inner object after
    # the declaration type, so the public type stays the same for every
    # variant and the wire layer does the renaming.
    group_xsd = entry["variants"][0]["xsd"]
    for field in entry["shared"] + entry["optional"]:
        info = nested_entry(entry, field)
        if not info:
            continue
        sub = nested_name(kind, field)
        wire_sub = wire_nested_name(kind, field)
        helper = lower + sub
        inner = inner_fields(info)
        inner_name = wire_sub + "Entry" if is_list(info) else wire_sub
        per_variant = nested_variants(entry, field)
        required = inner_presence(entry, field)

        mapping.append(f"// {helper} builds a nested object of a {kind.lower()} declaration.")
        if is_list(info):
            mapping.append(f"func {helper}(variant {variant}, value []{sub}) wire.{wire_sub}Payload {{")
            mapping.append("\tif len(value) == 0 {")
            mapping.append("\t\treturn nil")
            mapping.append("\t}")
        else:
            mapping.append(f"func {helper}(variant {variant}, value *{sub}) wire.{wire_sub}Payload {{")
            mapping.append("\tif value == nil {")
            mapping.append("\t\treturn nil")
            mapping.append("\t}")
        mapping.append("")
        mapping.append("\tswitch variant {")
        for v in entry["variants"]:
            variant_info = v["fields"].get(field)
            if not variant_info or not variant_info.get("nested"):
                continue
            name = SEMANTIC.get(v["code"], v["code"])
            key = variant_info["inner_key"]
            mapping.append(f"\tcase {name}:")
            if is_list(info):
                mapping.append(f"\t\tentries := make([]wire.{inner_name}, 0, len(value))")
                mapping.append("\t\tfor _, item := range value {")
                mapping.append(f"\t\t\tentries = append(entries, wire.{inner_name}{{")
                for inner_field, inner_info in per_variant[v["code"]].items():
                    innerkind, inner_field_name = field_kind(inner_field, group_xsd, inner_info, optional=inner_field not in required)
                    mapping.append(f"\t\t\t\t{inner_field_name}: {convert(innerkind).replace('{v}', 'item.' + inner_field_name)},")
                mapping.append("\t\t\t})")
                mapping.append("\t\t}")
                mapping.append(f"\t\treturn wire.{wire_sub}Payload{{")
                mapping.append(f'\t\t\t"{key}": entries,')
                mapping.append("\t\t}")
            else:
                mapping.append(f"\t\treturn wire.{wire_sub}Payload{{")
                mapping.append(f'\t\t\t"{key}": wire.{inner_name}{{')
                for inner_field, inner_info in per_variant[v["code"]].items():
                    innerkind, inner_field_name = field_kind(inner_field, group_xsd, inner_info, optional=inner_field not in required)
                    mapping.append(f"\t\t\t{inner_field_name}: {convert(innerkind).replace('{v}', 'value.' + inner_field_name)},")
                mapping.append("\t\t\t},")
                mapping.append("\t\t}")
        mapping.append("\tdefault:")
        mapping.append("\t\treturn nil")
        mapping.append("\t}")
        mapping.append("}")
        mapping.append("")

        mapping.append(f"// {helper}Of reads a nested object of a {kind.lower()} declaration.")
        if is_list(info):
            mapping.append(f"func {helper}Of(payload wire.{wire_sub}Payload) []{sub} {{")
            mapping.append("\tfor _, entries := range payload {")
            mapping.append("\t\tvalue := make([]" + sub + ", 0, len(entries))")
            mapping.append("\t\tfor _, entry := range entries {")
            mapping.append("\t\t\titem := " + sub + "{}")
            for inner_field, inner_info in nested_union_fields(entry, field):
                innerkind, inner_field_name = field_kind(inner_field, group_xsd, inner_info, optional=inner_field not in required)
                mapping.append(f"\t\t\titem.{inner_field_name} = {convert_back(innerkind).replace('{v}', 'entry.' + inner_field_name)}")
            mapping.append("\t\t\tvalue = append(value, item)")
            mapping.append("\t\t}")
            mapping.append("\t\treturn value")
            mapping.append("\t}")
            mapping.append("\treturn nil")
            mapping.append("}")
            mapping.append("")
        else:
            mapping.append(f"func {helper}Of(payload wire.{wire_sub}Payload) *{sub} {{")
            mapping.append("\tfor _, entry := range payload {")
            mapping.append("\t\tvalue := &" + sub + "{}")
            for inner_field, inner_info in nested_union_fields(entry, field):
                innerkind, inner_field_name = field_kind(inner_field, group_xsd, inner_info, optional=inner_field not in required)
                mapping.append(f"\t\tvalue.{inner_field_name} = {convert_back(innerkind).replace('{v}', 'entry.' + inner_field_name)}")
            mapping.append("\t\treturn value")
            mapping.append("\t}")
            mapping.append("\treturn nil")
            mapping.append("}")
            mapping.append("")

    mapping.append(f"// {variant} is a type of the {kind.lower()} declaration.")
    mapping.append(f"type {variant} string")
    mapping.append("")
    mapping.append("// The declarations of the family.")
    mapping.append("const (")
    for v in entry["variants"]:
        name = SEMANTIC.get(v["code"], v["code"])
        mapping.append(f'\t// {name} is the {v["family"]} declaration.')
        mapping.append(f'\t{name} {variant} = "{v["code"]}"')
    mapping.append(")")
    mapping.append("")

    mapping.append(f"// {kind[0].lower() + kind[1:]}EnvelopeKeys returns the root and the inner key of a declaration.")
    mapping.append(f"func {kind[0].lower() + kind[1:]}EnvelopeKeys(variant {variant}) (string, string) {{")
    mapping.append("\tswitch variant {")
    for v in entry["variants"]:
        name = SEMANTIC.get(v["code"], v["code"])
        mapping.append(f'\tcase {name}:')
        mapping.append(f'\t\treturn "{v["root"]}", "{v["inner"]}"')
    mapping.append("\tdefault:")
    mapping.append('\t\treturn "", ""')
    mapping.append("\t}")
    mapping.append("}")
    mapping.append("")

    mapping.append(f"// {kind[0].lower() + kind[1:]}Envelope wraps the fields of a declaration the way the API expects them.")
    mapping.append(f"func {kind[0].lower() + kind[1:]}Envelope(variant {variant}, fields map[string]any) map[string]any {{")
    mapping.append(f"\troot, inner := {kind[0].lower() + kind[1:]}EnvelopeKeys(variant)")
    mapping.append("\treturn map[string]any{")
    mapping.append("\t\troot: map[string]any{")
    mapping.append("\t\t\tinner: []any{fields},")
    mapping.append("\t\t},")
    mapping.append("\t}")
    mapping.append("}")
    mapping.append("")

    mapping.append(f"// {kind[0].lower() + kind[1:]}FieldsNotCarried lists the fields that a declaration type does not carry.")
    mapping.append(f"var {kind[0].lower() + kind[1:]}FieldsNotCarried = map[string][]string{{")
    for v in entry["variants"]:
        missing = [f for f in entry["optional"] if v["family"] not in entry["owner"][f]]
        if not missing:
            continue
        mapping.append(f'\t"{v["code"]}": {{')
        for field in missing:
            mapping.append(f'\t\t"{field}",')
        mapping.append("\t},")
    mapping.append("}")
    mapping.append("")

    mapping.append(f"// {method} submits a {kind.lower()} declaration.")
    mapping.append(f"func (c *Client) {method}(ctx context.Context, variant {variant}, declaration {struct}) ([]Submission, error) {{")
    mapping.append(f"\tfields, err := {kind[0].lower() + kind[1:]}FieldsOf(variant, declaration)")
    mapping.append("\tif err != nil {")
    mapping.append("\t\treturn nil, err")
    mapping.append("\t}")
    mapping.append("")
    mapping.append(f"\treturn c.submit(ctx, DocumentCode(variant), {kind[0].lower() + kind[1:]}Envelope(variant, fields))")
    mapping.append("}")
    mapping.append("")

    mapping.append(f"// {kind[0].lower() + kind[1:]}FieldsOf turns a {kind.lower()} declaration into the fields of the payload.")
    mapping.append(f"func {kind[0].lower() + kind[1:]}FieldsOf(variant {variant}, declaration {struct}) (map[string]any, error) {{")
    mapping.append(f"\tpayload := {payload_of}(declaration)")
    for field in entry["shared"] + entry["optional"]:
        info = entry["variants"][0]["fields"].get(field)
        fieldkind, name = field_kind(field, entry["variants"][0]["xsd"], info)
        if not fieldkind.startswith("nested:"):
            continue
        sub = nested_name(kind, field)
        helper = lower + sub
        mapping.append(f"\tpayload.{name} = {helper}(variant, declaration.{name})")
    mapping.append("")
    mapping.append("\tfields, err := declarationFields(payload)")
    mapping.append("\tif err != nil {")
    mapping.append("\t\treturn nil, err")
    mapping.append("\t}")
    mapping.append(f"\tdropFields(fields, {kind[0].lower() + kind[1:]}FieldsNotCarried[string(variant)])")
    mapping.append("")
    mapping.append("\treturn fields, nil")
    mapping.append("}")
    mapping.append("")

# The names of the public API must be unique.
names = set()
for group in GROUPS:
    for family, code, sample, xsd in group["variants"]:
        name = SEMANTIC.get(code, code)
        if name in names:
            raise SystemExit(f"the name {name} is used twice")
        names.add(name)
for (kind, field), name in NESTED_NAMES.items():
    if name in names:
        raise SystemExit(f"the name {name} is used twice")
    names.add(name)

os.makedirs(os.path.join(OUT, "internal", "wire"), exist_ok=True)
with open(os.path.join(OUT, "internal", "wire", "declarations.go"), "w", encoding="utf-8") as fh:
    fh.write("\n".join(wire))

with open(os.path.join(OUT, "declarations.go"), "w", encoding="utf-8") as fh:
    fh.write("\n".join(public))

# The translation lives in a file of its own: it is the glue between the
# public models and the wire format.
translation = []
translation.append('import (')
translation.append('\t"context"')
translation.append('\t"encoding/base64"')
translation.append('\t"encoding/json"')
translation.append('\t"fmt"')
if "date" in used_opt or "datetime" in used_opt or "time" in used_opt:
    translation.append('\t"time"')
translation.append("")
translation.append('\t"github.com/takispanag/ergani-go-sdk/internal/wire"')
translation.append(")")
translation.append("")
translation.append("// declarationFields turns a payload into the map of its fields, so that the")
translation.append("// fields a declaration type does not carry can be left out.")
translation.append("func declarationFields(payload any) (map[string]any, error) {")
translation.append("\traw, err := json.Marshal(payload)")
translation.append("\tif err != nil {")
translation.append('\t\treturn nil, fmt.Errorf("ergani: failed to marshal the declaration: %w", err)')
translation.append("\t}")
translation.append("")
translation.append("\tfields := map[string]any{}")
translation.append("\tif err := json.Unmarshal(raw, &fields); err != nil {")
translation.append('\t\treturn nil, fmt.Errorf("ergani: failed to decode the declaration: %w", err)')
translation.append("\t}")
translation.append("")
translation.append("\treturn fields, nil")
translation.append("}")
translation.append("")
translation.append("// dropFields removes the named fields from a payload.")
translation.append("func dropFields(fields map[string]any, names []string) {")
translation.append("\tfor _, name := range names {")
translation.append("\t\tdelete(fields, name)")
translation.append("\t}")
translation.append("}")
translation.append("")
translation.append("// boolPtr turns the API encoding of an optional boolean into a pointer.")
translation.append("//")
translation.append("// The API leaves a boolean empty when it is not answered, so the three")
translation.append("// states are kept apart.")
translation.append("func boolPtr(value wire.Bool) *bool {")
translation.append("\tif value == \"\" {")
translation.append("\t\treturn nil")
translation.append("\t}")
translation.append("")
translation.append("\tanswer := value == wire.True")
translation.append("")
translation.append("\treturn &answer")
translation.append("}")
translation.append("")
translation.append("// base64Bytes decodes a document of the API.")
translation.append("//")
translation.append("// A document that cannot be decoded yields no bytes rather than an error:")
translation.append("// the API has published a reference payload with a document that is not")
translation.append("// valid Base64, and a client must not fail on a document it only passes")
translation.append("// through.")
translation.append("func base64Bytes(value wire.Base64) []byte {")
translation.append("\tdecoded, err := base64.StdEncoding.DecodeString(string(value))")
translation.append("\tif err != nil {")
translation.append("\t\treturn nil")
translation.append("\t}")
translation.append("")
translation.append("\treturn decoded")
translation.append("}")
translation.append("")
translation.extend(mapping)

# ---------------------------------------------------------------------------
# The golden test.
# ---------------------------------------------------------------------------

test = []
test.append("// Code generated by tools/generate_declarations.py. DO NOT EDIT.")
test.append("")
test.append("package ergani")
test.append("")
test.append("import (")
test.append('\t"encoding/json"')
test.append('\t"os"')
test.append('\t"path/filepath"')
test.append('\t"strings"')
test.append('\t"testing"')
test.append("")
test.append('\t"github.com/takispanag/ergani-go-sdk/internal/wire"')
test.append(")")
test.append("")
test.append("// declarationSample is a reference payload of the Ministry and the way to")
test.append("// turn it back into the payload of the SDK.")
test.append("type declarationSample struct {")
test.append("\t// name is the code of the declaration type.")
test.append("\tname string")
test.append("")
test.append("\t// file is the name of the reference payload.")
test.append("\tfile string")
test.append("")
test.append("\t// roundTrip turns the fields of the reference payload into the fields of")
test.append("\t// the payload of the SDK.")
test.append("\troundTrip func(map[string]any) (map[string]any, error)")
test.append("}")
test.append("")
test.append("// declarationSamples are the reference payloads of every declaration type the")
test.append("// SDK models.")
test.append("var declarationSamples = []declarationSample{")
for entry in groups:
    group = entry["group"]
    kind = group["kind"]
    lower = kind[0].lower() + kind[1:]
    for v in entry["variants"]:
        name = SEMANTIC.get(v["code"], v["code"])
        test.append("\t{")
        test.append(f'\t\tname: "{v["code"]}",')
        test.append(f'\t\tfile: "{v["sample"]}",')
        test.append("\t\troundTrip: func(fields map[string]any) (map[string]any, error) {")
        test.append("\t\t\traw, err := json.Marshal(fields)")
        test.append("\t\t\tif err != nil {")
        test.append("\t\t\t\treturn nil, err")
        test.append("\t\t\t}")
        test.append("")
        test.append(f"\t\t\tvar payload wire.{group['wire']}")
        test.append("\t\t\tif err := json.Unmarshal(raw, &payload); err != nil {")
        test.append("\t\t\t\treturn nil, err")
        test.append("\t\t\t}")
        test.append("")
        test.append(f"\t\t\treturn {lower}FieldsOf({name}, {lower}DeclarationOf(payload))")
        test.append("\t\t},")
        test.append("\t},")
test.append("}")
test.append("")
test.append("// TestDeclarationsMatchTheOfficialSamples checks the payload of every")
test.append("// declaration type against the reference payload of the Ministry.")
test.append("func TestDeclarationsMatchTheOfficialSamples(t *testing.T) {")
test.append("\tfor _, sample := range declarationSamples {")
test.append("\t\tt.Run(sample.name, func(t *testing.T) {")
test.append("\t\t\twant := readDeclarationSample(t, sample.file)")
test.append("")
test.append("\t\t\tgot, err := sample.roundTrip(want)")
test.append("\t\t\tif err != nil {")
test.append('\t\t\t\tt.Fatalf("the payload of %s could not be built: %v", sample.name, err)')
test.append("\t\t\t}")
test.append("")
test.append("\t\t\tif diff := compareFields(got, want); len(diff) > 0 {")
test.append('\t\t\t\tt.Errorf("the payload of %s does not match the official sample:\\n  %s",')
test.append("\t\t\t\t\tsample.name, strings.Join(diff, \"\\n  \"))")
test.append("\t\t\t}")
test.append("\t\t})")
test.append("\t}")
test.append("}")
test.append("")
test.append("// readDeclarationSample reads a reference payload and returns the fields of")
test.append("// the declaration.")
test.append("func readDeclarationSample(t *testing.T, name string) map[string]any {")
test.append("\tt.Helper()")
test.append("")
test.append('\tpath := filepath.Join("testdata", "declarations", name)')
test.append("\tdata, err := os.ReadFile(path)")
test.append("\tif err != nil {")
test.append('\t\tt.Fatalf("Failed to read %s: %v", path, err)')
test.append("\t}")
test.append("")
test.append("\tvar document map[string]any")
test.append("\tif err := json.Unmarshal(data, &document); err != nil {")
test.append('\t\tt.Fatalf("Failed to decode %s: %v", path, err)')
test.append("\t}")
test.append("")
test.append("\tfor _, value := range document {")
test.append("\t\tobject, ok := value.(map[string]any)")
test.append("\t\tif !ok {")
test.append('\t\t\tt.Fatalf("The root of %s does not hold an object", path)')
test.append("\t\t}")
test.append("\t\tfor _, inner := range object {")
test.append("\t\t\tif list, ok := inner.([]any); ok {")
test.append("\t\t\t\tif len(list) == 0 {")
test.append('\t\t\t\t\tt.Fatalf("The declaration of %s is empty", path)')
test.append("\t\t\t\t}")
test.append("\t\t\t\tinner = list[0]")
test.append("\t\t\t}")
test.append("\t\t\tfields, ok := inner.(map[string]any)")
test.append("\t\t\tif !ok {")
test.append('\t\t\t\tt.Fatalf("The declaration of %s is not an object", path)')
test.append("\t\t\t}")
test.append("\t\t\treturn fields")
test.append("\t\t}")
test.append("\t}")
test.append("")
test.append('\tt.Fatalf("The declaration of %s is empty", path)')
test.append("")
test.append("\treturn nil")
test.append("}")
test.append("")

with open(os.path.join(OUT, "declarations_golden_test.go"), "w", encoding="utf-8") as fh:
    fh.write("\n".join(test))

# The helpers of the optional values are computed before the translation is
# written, because the translation uses them.
# ---------------------------------------------------------------------------
# The helpers of the optional values.
# ---------------------------------------------------------------------------

helpers = []

emitted = set()
for kind in sorted(used_opt):
    if kind == "flag":
        # A boolean carries the three states on its own.
        helpers.append("// optBoolPtr converts an optional boolean into the wire encoding.")
        helpers.append("func optBoolPtr(value *bool) *wire.Bool {")
        helpers.append("\tif value == nil {")
        helpers.append("\t\treturn nil")
        helpers.append("\t}")
        helpers.append("\tanswer := wire.FromBool(*value)")
        helpers.append("\treturn &answer")
        helpers.append("}")
        helpers.append("")
        helpers.append("// boolPtrOf reads an optional boolean of the wire encoding.")
        helpers.append("func boolPtrOf(value *wire.Bool) *bool {")
        helpers.append("\tif value == nil {")
        helpers.append("\t\treturn nil")
        helpers.append("\t}")
        helpers.append("\tanswer := *value == wire.True")
        helpers.append("\treturn &answer")
        helpers.append("}")
        helpers.append("")
        continue

    public, wire_type, to_wire, from_wire = OPT_BODIES[kind]
    name, back = OPT[kind]

    if name in emitted:
        continue
    emitted.add(name)
    emitted.add(back)

    helpers.append(f"// {name} converts an optional value into the wire encoding.")
    helpers.append(f"func {name}(value {public}) {wire_type} {{")
    helpers.append("\tif value == nil {")
    helpers.append("\t\treturn nil")
    helpers.append("\t}")
    helpers.append(f"\t{to_wire}")
    helpers.append("}")
    helpers.append("")

    helpers.append(f"// {back} reads an optional value of the wire encoding.")
    helpers.append(f"func {back}(value {wire_type}) {public} {{")
    helpers.append("\tif value == nil {")
    helpers.append("\t\treturn nil")
    helpers.append("\t}")
    helpers.append(f"\t{from_wire}")
    helpers.append("}")
    helpers.append("")

# The translation is written with the helpers of the optional values in it.
header = []
header.append("// Code generated by tools/generate_declarations.py. DO NOT EDIT.")
header.append("")
header.append("package ergani")
header.append("")

with open(os.path.join(OUT, "declarations_wire.go"), "w", encoding="utf-8") as fh:
    fh.write("\n".join(header + translation + helpers))

print("wrote the wire structs, the public models and the mapping")
for entry in groups:
    print(f"  {entry['group']['kind']:20s} {len(entry['shared']):3d} shared, {len(entry['optional']):3d} optional, {len(entry['variants'])} variants")
