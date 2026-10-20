package ergani

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/takispanag/ergani-go-sdk/internal/wire"
)

// HireVariant is a type of the notification of the start of an employment.
//
// The value is the code of the declaration type, as the API expects it in the
// path of the submission.
type HireVariant string

// The notifications of the start of an employment.
const (
	// HireNew is the notification of a hire.
	HireNew HireVariant = "WebE3N"

	// HireTransfer is the notification of the start of an employment with a
	// transfer from another company.
	HireTransfer HireVariant = "WebE3M"

	// HireBorrowed is the notification of the start of an employment with
	// staff borrowed from another company.
	HireBorrowed HireVariant = "WebE3D"

	// HireForLoanOut is the notification of the hire of an employee who is
	// loaned out to another company.
	HireForLoanOut HireVariant = "WebE3PD"
)

// TerminationVariant is a type of the notification of the end of an
// employment.
//
// The value is the code of the declaration type, as the API expects it in the
// path of the submission.
type TerminationVariant string

// The notifications of the end of an employment.
const (
	// TerminationResignation is the notification of a voluntary resignation.
	TerminationResignation TerminationVariant = "WebE5N"

	// TerminationNotice is the notification of the employer that the
	// employee was pressed to resign.
	TerminationNotice TerminationVariant = "WebE5O"

	// TerminationResignationAfterNotice is the notification of a resignation
	// that follows a notice.
	TerminationResignationAfterNotice TerminationVariant = "WebE5AO"

	// TerminationVoluntaryExit is the notification of a voluntary exit.
	TerminationVoluntaryExit TerminationVariant = "WebE5E"

	// TerminationRetirement is the notification of a retirement.
	TerminationRetirement TerminationVariant = "WebE5S"

	// TerminationFifteenYears is the notification of a resignation after
	// fifteen years of service or upon reaching the retirement age.
	TerminationFifteenYears TerminationVariant = "WebE5DS"

	// TerminationDeath is the notification of the end of an employment
	// because of the death of the employee.
	TerminationDeath TerminationVariant = "WebE5D"

	// TerminationDismissalWithoutNotice is the notification of a dismissal
	// without notice.
	TerminationDismissalWithoutNotice TerminationVariant = "WebE6NXP"

	// TerminationDismissalWithNotice is the notification of a dismissal with
	// notice.
	TerminationDismissalWithNotice TerminationVariant = "WebE6NMP"

	// TerminationRetirementByDismissal is the notification of a retirement by
	// dismissal without notice.
	TerminationRetirementByDismissal TerminationVariant = "WebE6SXP"

	// TerminationProbation is the notification of the automatic termination
	// of a probation period.
	TerminationProbation TerminationVariant = "WebE6LT"

	// TerminationTransfer is the notification of the end of an employment
	// because the employee moves to another company.
	TerminationTransfer TerminationVariant = "WebE6M"

	// TerminationFixedTerm is the notification of the end of a fixed term
	// contract.
	TerminationFixedTerm TerminationVariant = "WebE7N"
)

// SubmitHire submits the notification of the start of an employment.
//
// The variant selects the type of the declaration, and with it the fields
// the API expects. The fields the variant does not carry are left out of the
// payload.
//
// The declaration carries the fields of the employee, of the employment and
// of the branch. Every field is documented with the official description of
// the API.
//
// It returns a [*ValidationError] when the declaration is malformed, and an
// [*APIError] when the API rejects it.
//
// The submission is not idempotent. When the API answers with a transient
// failure after it has recorded the declaration, the retry may record it a
// second time.
func (c *Client) SubmitHire(ctx context.Context, variant HireVariant, declaration HireDeclaration) ([]Submission, error) {
	fields, err := hireFieldsOf(variant, declaration)
	if err != nil {
		return nil, err
	}

	return c.submit(ctx, DocumentCode(variant), employmentEnvelope(string(variant), fields))
}

// SubmitTermination submits the notification of the end of an employment.
//
// The variant selects the type of the declaration, and with it the fields
// the API expects. The fields the variant does not carry are left out of the
// payload.
//
// It returns a [*ValidationError] when the declaration is malformed, and an
// [*APIError] when the API rejects it.
//
// The submission is not idempotent. When the API answers with a transient
// failure after it has recorded the declaration, the retry may record it a
// second time.
func (c *Client) SubmitTermination(ctx context.Context, variant TerminationVariant, declaration TerminationDeclaration) ([]Submission, error) {
	fields, err := terminationFieldsOf(variant, declaration)
	if err != nil {
		return nil, err
	}

	return c.submit(ctx, DocumentCode(variant), employmentEnvelope(string(variant), fields))
}

// hireFieldsOf turns a hire declaration into the fields of the payload of the
// API, for the variant of the declaration.
func hireFieldsOf(variant HireVariant, declaration HireDeclaration) (map[string]any, error) {
	payload := hirePayloadOf(declaration)
	payload.BeneficiarySelections = beneficiarySelections(variant, declaration.BeneficiarySelections.Code)

	fields, err := declarationFields(payload)
	if err != nil {
		return nil, err
	}
	dropFields(fields, hireFieldsNotCarried[string(variant)])

	return fields, nil
}

// terminationFieldsOf turns a termination declaration into the fields of the
// payload of the API, for the variant of the declaration.
func terminationFieldsOf(variant TerminationVariant, declaration TerminationDeclaration) (map[string]any, error) {
	fields, err := declarationFields(terminationPayloadOf(declaration))
	if err != nil {
		return nil, err
	}
	dropFields(fields, terminationFieldsNotCarried[string(variant)])

	return fields, nil
}

// employmentEnvelope wraps the fields of a declaration the way the API
// expects them: a root, an object, and the declarations themselves.
//
// The declarations are a list. The schemas of the API declare the inner
// element with maxOccurs="unbounded", so a single declaration is a list of
// one.
func employmentEnvelope(variant string, fields map[string]any) map[string]any {
	return map[string]any{
		"Anaggelies" + variant: map[string]any{
			"Anaggelia" + variant: []any{fields},
		},
	}
}

// family returns the name of the schema of the variant, which the API uses
// to name the objects it nests.
func (v HireVariant) family() string {
	switch v {
	case HireNew:
		return "E3N"
	case HireTransfer:
		return "E3M"
	case HireBorrowed:
		return "E3D"
	case HireForLoanOut:
		return "E3PD"
	default:
		return string(v)
	}
}

// beneficiarySelections builds the selection of the beneficiary funds.
//
// The API names the inner object after the schema of the declaration, so
// the selection of a hire is "EpikourikiSelectionsE3N" and the one of a
// transfer is "EpikourikiSelectionsE3M".
func beneficiarySelections(variant HireVariant, code string) map[string]wire.BeneficiarySelection {
	if code == "" {
		return nil
	}

	return map[string]wire.BeneficiarySelection{
		"EpikourikiSelections" + variant.family(): {Code: code},
	}
}

// beneficiaryCode returns the code of the selection of the beneficiary funds.
func beneficiaryCode(selections map[string]wire.BeneficiarySelection) string {
	for _, selection := range selections {
		return selection.Code
	}

	return ""
}

// boolPtr turns the API encoding of an optional boolean into a pointer.
//
// The API leaves a boolean empty when it is not answered, so the three states
// are kept apart.
func boolPtr(value wire.Bool) *bool {
	if value == "" {
		return nil
	}

	answer := value == wire.True

	return &answer
}

// base64Bytes decodes a document of the API.
//
// A document that cannot be decoded yields no bytes rather than an error:
// the API has published a reference payload with a document that is not
// valid Base64, and a client must not fail on a document it only passes
// through.
func base64Bytes(value wire.Base64) []byte {
	decoded, err := base64.StdEncoding.DecodeString(string(value))
	if err != nil {
		return nil
	}

	return decoded
}

// declarationFields turns a payload into the map of its fields, so that the
// fields a declaration type does not carry can be left out.
func declarationFields(payload any) (map[string]any, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("ergani: failed to marshal the declaration: %w", err)
	}

	fields := map[string]any{}
	if err := json.Unmarshal(raw, &fields); err != nil {
		return nil, fmt.Errorf("ergani: failed to decode the declaration: %w", err)
	}

	return fields, nil
}

// dropFields removes the named fields from a payload.
func dropFields(fields map[string]any, names []string) {
	for _, name := range names {
		delete(fields, name)
	}
}
