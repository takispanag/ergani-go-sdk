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
    "minutes": "int",
    "flag": "*bool",
    "base64": "[]byte",
    "enum": "string",
    "nested": "json.RawMessage",
}

WIRE_TYPE = {
    "string": "string",
    "date": "wire.Date",
    "datetime": "wire.DateTime",
    "time": "wire.Time",
    "amount": "wire.Amount",
    "hours": "wire.Hours",
    "count": "wire.Count",
    "minutes": "wire.Minutes",
    "flag": "wire.Bool",
    "base64": "wire.Base64",
    "enum": "string",
    "nested": "json.RawMessage",
}

CONVERT = {
    "string": "{v}",
    "date": "wire.Date{Time: {v}}",
    "datetime": "wire.DateTime{Time: {v}}",
    "time": "wire.Time{Time: {v}}",
    "amount": "wire.Amount({v})",
    "hours": "wire.Hours({v})",
    "count": "wire.Count({v})",
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
    "minutes": "int({v})",
    "flag": "boolPtr({v})",
    "base64": "base64Bytes({v})",
    "enum": "string({v})",
    "nested": "{v}",
}

ENUM_PREFIX = "enum:"


def public_type(kind):
    if kind.startswith(ENUM_PREFIX):
        return kind[len(ENUM_PREFIX):]
    return PUBLIC_TYPE[kind]


def wire_type(kind):
    if kind.startswith(ENUM_PREFIX):
        return "string"
    return WIRE_TYPE[kind]


def wire_type_local(kind):
    if kind.startswith(ENUM_PREFIX):
        return "string"
    return WIRE_TYPE[kind].replace("wire.", "")


def convert(kind):
    if kind.startswith(ENUM_PREFIX):
        return "string({v})"
    return CONVERT[kind]


def convert_back(kind):
    if kind.startswith(ENUM_PREFIX):
        return f"{kind[len(ENUM_PREFIX):]}({{v}})"
    return CONVERT_BACK[kind]


def go_name(field):
    """The name of a field in the public API, when it is not known."""
    parts = field[2:].split("_")
    name = ""
    for part in parts:
        if not part:
            continue
        name += INITIALISMS.get(part, part[:1].upper() + part[1:])
    return name


def field_kind(field, xsd, value=None):
    """The type of a field, and its name in the public API."""
    if field in KNOWN:
        return KNOWN[field]

    if isinstance(value, (dict, list)):
        return ("nested", go_name(field))

    kind = SCHEMA_TYPES.get(xsd, {}).get(field, "string")

    return (kind, go_name(field))


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

    The fields are the ones of the declaration itself. The nested objects are
    kept whole, because the API names their inner object after the variant.
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

    return {key: item for key, item in value.items()}, root, inner_key


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
wire.append('import "encoding/json"')
wire.append("")

for entry in groups:
    group = entry["group"]
    wire.append(f"// {group['wire']} is the flat set of the fields of a {group['kind'].lower()} declaration.")
    wire.append(f"type {group['wire']} struct {{")
    for field in entry["shared"] + entry["optional"]:
        kind, name = field_kind(field, entry["variants"][0]["xsd"], entry["variants"][0]["fields"].get(field))
        wire.append(f'\t{name} {wire_type_local(kind)} `json:"{field}"`')
    wire.append("}")
    wire.append("")

# ---------------------------------------------------------------------------
# The public models.
# ---------------------------------------------------------------------------

public = []
public.append("// Code generated by tools/generate_declarations.py. DO NOT EDIT.")
public.append("")
public.append("package ergani")
public.append("")
public.append('import "encoding/json"')
public.append('import "time"')
public.append("")

for entry in groups:
    group = entry["group"]
    public.append(f"// {group['struct']} is the declaration of the {group['kind'].lower()} of an employer.")
    public.append(f"type {group['struct']} struct {{")
    for field in entry["shared"] + entry["optional"]:
        kind, name = field_kind(field, entry["variants"][0]["xsd"], entry["variants"][0]["fields"].get(field))
        doc = DOCS.get(field, "")
        if doc:
            public.append(f"\t// {name} — {doc}")
        public.append(f"\t{name} {public_type(kind)}")
    public.append("}")
    public.append("")

# ---------------------------------------------------------------------------
# The translation and the submission.
# ---------------------------------------------------------------------------

mapping = []
mapping.append("// Code generated by tools/generate_declarations.py. DO NOT EDIT.")
mapping.append("")
mapping.append("package ergani")
mapping.append("")
mapping.append('import "context"')
mapping.append('import "github.com/takispanag/ergani-go-sdk/internal/wire"')
mapping.append("")

for entry in groups:
    group = entry["group"]
    kind = group["kind"]
    struct = group["struct"]
    wire_name = group["wire"]
    variant = group["variant"]
    method = group["method"]
    payload_of = kind[0].lower() + kind[1:] + "PayloadOf"
    declaration_of = kind[0].lower() + kind[1:] + "DeclarationOf"

    mapping.append(f"// {payload_of} turns a {struct} into the payload of the API.")
    mapping.append(f"func {payload_of}(d {struct}) wire.{wire_name} {{")
    mapping.append(f"\tp := wire.{wire_name}{{}}")
    for field in entry["shared"] + entry["optional"]:
        fieldkind, name = field_kind(field, entry["variants"][0]["xsd"], entry["variants"][0]["fields"].get(field))
        mapping.append(f"\tp.{name} = " + convert(fieldkind).replace("{v}", f"d.{name}"))
    mapping.append("\treturn p")
    mapping.append("}")
    mapping.append("")

    mapping.append(f"// {declaration_of} turns a payload of the API into a {struct}.")
    mapping.append(f"func {declaration_of}(p wire.{wire_name}) {struct} {{")
    mapping.append(f"\tvar d {struct}")
    for field in entry["shared"] + entry["optional"]:
        fieldkind, name = field_kind(field, entry["variants"][0]["xsd"], entry["variants"][0]["fields"].get(field))
        mapping.append(f"\td.{name} = " + convert_back(fieldkind).replace("{v}", f"p.{name}"))
    mapping.append("\treturn d")
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

os.makedirs(os.path.join(OUT, "internal", "wire"), exist_ok=True)
with open(os.path.join(OUT, "internal", "wire", "declarations.go"), "w", encoding="utf-8") as fh:
    fh.write("\n".join(wire))

with open(os.path.join(OUT, "declarations.go"), "w", encoding="utf-8") as fh:
    fh.write("\n".join(public))

with open(os.path.join(OUT, "declarations_map.go"), "w", encoding="utf-8") as fh:
    fh.write("\n".join(mapping))

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

print("wrote the wire structs, the public models and the mapping")
for entry in groups:
    print(f"  {entry['group']['kind']:20s} {len(entry['shared']):3d} shared, {len(entry['optional']):3d} optional, {len(entry['variants'])} variants")
