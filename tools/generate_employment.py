"""Generate the Go code of the employment-lifecycle declarations.

One table drives the wire structs, the public models and the translation
between them, so the three cannot drift apart. The generator is committed
under tools/ and the code it writes is committed too, so the build does not
need the generator.
"""

import json
import os
import re
from collections import OrderedDict

SRC = r"C:\Users\Takis\AppData\Local\Temp\opencode\ergani-docs\json"
OUT = r"C:\Users\Takis\AppData\Local\Temp\opencode\ergani-go-sdk"
DOCS_FILE = r"C:\Users\Takis\AppData\Local\Temp\opencode\field-docs.json"

# The official Greek description of every field, taken from the XSD schemas.
with open(DOCS_FILE, encoding="utf-8") as fh:
    DOCS = json.load(fh)

# The type of every field, and the name of the field in the public API.
# date and time are time.Time, amount and hours are float64, count and
# minutes are int, bool is a bool, base64 is a []byte, string is a string,
# and enum:Name is one of the enumerated types of the package.
TYPES = {
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

# The nested object of the beneficiary selections. It carries one entry per
# variant of the declaration, because the API names the inner object after it.
NESTED = {
    "EpikourikiSelections": ("BeneficiarySelections", "f_kod_epikourikis", "Code", ["E3N", "E3M", "E3PD"]),
}

FAMILIES = OrderedDict(
    [
        ("E3N", "Hire"),
        ("E3M", "Hire"),
        ("E3D", "Hire"),
        ("E3PD", "Hire"),
        ("E5N", "Termination"),
        ("E5O", "Termination"),
        ("E5AO", "Termination"),
        ("E5E", "Termination"),
        ("E5S", "Termination"),
        ("E5DS", "Termination"),
        ("E5D", "Termination"),
        ("E6NXP", "Termination"),
        ("E6NMP", "Termination"),
        ("E6SXP", "Termination"),
        ("E6LT", "Termination"),
        ("E6M", "Termination"),
        ("E7N", "Termination"),
    ]
)

SAMPLES = {
    "E3N": "testProslipsiNew.json",
    "E3M": "testMetabibashFrom.json",
    "E3D": "testDaneismosFrom.json",
    "E3PD": "testProslipsiDaneismos.json",
    "E5N": "testApoxwrhshsNew.json",
    "E5O": "testApoxwrhshOxlhsh.json",
    "E5AO": "testApoxwrhshsAfterOxlhsh.json",
    "E5E": "testApoxwrhshEthelousia.json",
    "E5S": "testApoxwrhshsSyntaksh.json",
    "E5DS": "testSyntakshDekapentaetia.json",
    "E5D": "testApoxwrhshsThanatos.json",
    "E6NXP": "testKataggeliaXwrisProidopoihsh.json",
    "E6NMP": "testKataggeliaMeProidopoihsh.json",
    "E6SXP": "testSyntakshXwrisProidopoihsh.json",
    "E6LT": "testLyshsDokimastikhs.json",
    "E6M": "testMetabibashTo.json",
    "E7N": "testLyshSymbashsNew.json",
}

PUBLIC_TYPE = {
    "string": "string",
    "date": "time.Time",
    "time": "time.Time",
    "amount": "float64",
    "hours": "float64",
    "count": "int",
    "minutes": "int",
    "flag": "*bool",
    "base64": "[]byte",
}

WIRE_TYPE = {
    "string": "string",
    "date": "wire.Date",
    "time": "wire.Time",
    "amount": "wire.Amount",
    "hours": "wire.Hours",
    "count": "wire.Count",
    "minutes": "wire.Minutes",
    "flag": "wire.Bool",
    "base64": "wire.Base64",
}

CONVERT = {
    "string": "{v}",
    "date": "wire.Date{Time: {v}}",
    "time": "wire.Time{Time: {v}}",
    "amount": "wire.Amount({v})",
    "hours": "wire.Hours({v})",
    "count": "wire.Count({v})",
    "minutes": "wire.Minutes({v})",
    "flag": "wire.FromBoolPtr({v})",
    "base64": "wire.Base64FromBytes({v})",
}

CONVERT_BACK = {
    "string": "{v}",
    "date": "{v}.Time",
    "time": "{v}.Time",
    "amount": "float64({v})",
    "hours": "float64({v})",
    "count": "int({v})",
    "minutes": "int({v})",
    "flag": "boolPtr({v})",
    "base64": "base64Bytes({v})",
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
    with open(os.path.join(SRC, filename), encoding="utf-8-sig") as fh:
        data = json.load(fh)
    found = {}
    leaves(data, "", found)
    trimmed = {}
    for p, v in found.items():
        parts = p.split(".")
        if len(parts) < 3:
            continue
        trimmed[parts[2]] = v
    return trimmed


family_fields = {f: fields_of(SAMPLES[f]) for f in FAMILIES}

unions = {}
owner = {}
for family, kind in FAMILIES.items():
    for field in family_fields[family]:
        union = unions.setdefault(kind, [])
        if field not in union:
            union.append(field)
        owner.setdefault(field, set()).add(family)

shared = {}
optional = {}
for kind, union in unions.items():
    variants = [f for f in FAMILIES if FAMILIES[f] == kind]
    common = set.intersection(*(set(family_fields[f]) for f in variants))
    shared[kind] = [f for f in union if f in common]
    optional[kind] = [f for f in union if f not in common]

# ---------------------------------------------------------------------------
# The wire structs.
# ---------------------------------------------------------------------------

wire = []
wire.append("// Code generated by tools/generate_employment.go. DO NOT EDIT.")
wire.append("")
wire.append("package wire")
wire.append("")

for kind in unions:
    name = "HirePayload" if kind == "Hire" else "TerminationPayload"
    wire.append(f"// {name} is the flat set of the fields of a {kind.lower()} declaration.")
    wire.append(f"type {name} struct {{")
    for field in shared[kind] + optional[kind]:
        if field in NESTED:
            continue
        kindof, goname = TYPES[field]
        wire.append(f'\t{goname} {wire_type_local(kindof)} `json:"{field}"`')
    for field, (goname, wirefield, subname, variants) in NESTED.items():
        if field in unions[kind]:
            wire.append(f'\t{goname} map[string]BeneficiarySelection `json:"{field}"`')
    wire.append("}")
    wire.append("")

for field, (goname, wirefield, subname, variants) in NESTED.items():
    wire.append("// BeneficiarySelection is one selection of a beneficiary fund.")
    wire.append("type BeneficiarySelection struct {")
    wire.append(f'\tCode string `json:"{wirefield}"`')
    wire.append("}")
    wire.append("")

# ---------------------------------------------------------------------------
# The public models.
# ---------------------------------------------------------------------------

public = []
public.append("// Code generated by tools/generate_employment.go. DO NOT EDIT.")
public.append("")
public.append("package ergani")
public.append("")
public.append('import "time"')
public.append("")

for kind in unions:
    name = "HireDeclaration" if kind == "Hire" else "TerminationDeclaration"
    public.append(f"// {name} is the declaration of the {'start' if kind == 'Hire' else 'end'} of an employment.")
    public.append(f"type {name} struct {{")
    for field in shared[kind] + optional[kind]:
        if field in NESTED:
            continue
        kindof, goname = TYPES[field]
        doc = DOCS.get(field, "")
        if doc:
            public.append(f"\t// {goname} — {doc}")
        public.append(f"\t{goname} {public_type(kindof)}")
    for field, (goname, wirefield, subname, variants) in NESTED.items():
        if field in unions[kind]:
            public.append(f"\t{goname} {goname}")
    public.append("}")
    public.append("")

for field, (goname, wirefield, subname, variants) in NESTED.items():
    public.append(f"// {goname} is the selection of the beneficiary funds.")
    public.append(f"type {goname} struct {{")
    public.append("\t// Code is the code of the selected fund.")
    public.append("\tCode string")
    public.append("}")
    public.append("")

# ---------------------------------------------------------------------------
# The translation.
# ---------------------------------------------------------------------------

mapping = []
mapping.append("// Code generated by tools/generate_employment.go. DO NOT EDIT.")
mapping.append("")
mapping.append("package ergani")
mapping.append("")
mapping.append('import "github.com/takispanag/ergani-go-sdk/internal/wire"')
mapping.append("")

for kind in unions:
    name = "HireDeclaration" if kind == "Hire" else "TerminationDeclaration"
    wire_name = "HirePayload" if kind == "Hire" else "TerminationPayload"
    fn = "hirePayloadOf" if kind == "Hire" else "terminationPayloadOf"
    back = "hireDeclarationOf" if kind == "Hire" else "terminationDeclarationOf"
    mapping.append(f"// {fn} turns a {name} into the payload of the API.")
    mapping.append(f"func {fn}(d {name}) wire.{wire_name} {{")
    mapping.append(f"\tp := wire.{wire_name}{{}}")
    for field in shared[kind] + optional[kind]:
        if field in NESTED:
            continue
        kindof, goname = TYPES[field]
        mapping.append(f"\tp.{goname} = " + convert(kindof).replace("{v}", f"d.{goname}"))
    mapping.append("\treturn p")
    mapping.append("}")
    mapping.append("")
    mapping.append(f"// {back} turns a payload of the API into a {name}.")
    mapping.append(f"func {back}(p wire.{wire_name}) {name} {{")
    mapping.append(f"\tvar d {name}")
    for field in shared[kind] + optional[kind]:
        if field in NESTED:
            continue
        kindof, goname = TYPES[field]
        mapping.append(f"\td.{goname} = " + convert_back(kindof).replace("{v}", f"p.{goname}"))
    for field, (goname, wirefield, subname, variants) in NESTED.items():
        if field in unions[kind]:
            mapping.append(f"\td.{goname}.Code = beneficiaryCode(p.{goname})")
    mapping.append("\treturn d")
    mapping.append("}")
    mapping.append("")

# The code of the declaration type of every family. The families are the
# names of the schemas, the codes are what the API expects in the path.
CODES = {
    "E3N": "WebE3N",
    "E3M": "WebE3M",
    "E3D": "WebE3D",
    "E3PD": "WebE3PD",
    "E5N": "WebE5N",
    "E5O": "WebE5O",
    "E5AO": "WebE5AO",
    "E5E": "WebE5E",
    "E5S": "WebE5S",
    "E5DS": "WebE5DS",
    "E5D": "WebE5D",
    "E6NXP": "WebE6NXP",
    "E6NMP": "WebE6NMP",
    "E6SXP": "WebE6SXP",
    "E6LT": "WebE6LT",
    "E6M": "WebE6M",
    "E7N": "WebE7N",
}

# The fields that each variant does not carry.
for kind in unions:
    fn = "hireFieldsNotCarried" if kind == "Hire" else "terminationFieldsNotCarried"
    mapping.append(f"// {fn} lists the fields that a declaration type does not carry.")
    mapping.append(f"var {fn} = map[string][]string{{")
    for family, fkind in FAMILIES.items():
        if fkind != kind:
            continue
        missing = [field for field in optional[kind] if family not in owner[field]]
        if not missing:
            continue
        mapping.append(f'\t"{CODES[family]}": {{')
        for field in missing:
            mapping.append(f'\t\t"{field}",')
        mapping.append("\t},")
    mapping.append("}")
    mapping.append("")

os.makedirs(os.path.join(OUT, "internal", "wire"), exist_ok=True)
with open(os.path.join(OUT, "internal", "wire", "employment.go"), "w", encoding="utf-8") as fh:
    fh.write("\n".join(wire))

with open(os.path.join(OUT, "employment.go"), "w", encoding="utf-8") as fh:
    fh.write("\n".join(public))

with open(os.path.join(OUT, "employment_map.go"), "w", encoding="utf-8") as fh:
    fh.write("\n".join(mapping))

print("wrote the wire struct, the public model and the mapping")
for kind in unions:
    print(f"  {kind}: {len(shared[kind])} shared, {len(optional[kind])} optional")
