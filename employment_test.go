package ergani

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/takispanag/ergani-go-sdk/internal/wire"
)

// The reference payloads of the employment lifecycle declarations, as the
// Ministry of Labour publishes them. The SDK must produce the same payload,
// field for field.

// The variants of the family and the reference payload of each.
var employmentSamples = []struct {
	variant string
	file    string
}{
	{"WebE3N", "E3N.json"},
	{"WebE3M", "E3M.json"},
	{"WebE3D", "E3D.json"},
	{"WebE3PD", "E3PD.json"},
	{"WebE5N", "E5N.json"},
	{"WebE5O", "E5O.json"},
	{"WebE5AO", "E5AO.json"},
	{"WebE5E", "E5E.json"},
	{"WebE5S", "E5S.json"},
	{"WebE5DS", "E5DS.json"},
	{"WebE5D", "E5D.json"},
	{"WebE6NXP", "E6NXP.json"},
	{"WebE6NMP", "E6NMP.json"},
	{"WebE6SXP", "E6SXP.json"},
	{"WebE6LT", "E6LT.json"},
	{"WebE6M", "E6M.json"},
	{"WebE7N", "E7N.json"},
}

// The fields the API encodes as a number. Their exact spelling varies between
// the samples, so they are compared by their value.
var numericFields = map[string]int{
	"f_apodoxes":                2,
	"f_hour_apodoxes":           2,
	"f_posoapozimiosis":         2,
	"f_week_hours":              1,
	"f_full_employment_hours":   1,
	"f_arithmos_teknon":         0,
	"f_dialeimma_minutes":       0,
	"f_euelikto_wrario_minutes": 0,
}

// TestEmploymentDeclarationsMatchTheOfficialSamples checks the payload of
// every variant of the employment lifecycle against the reference payload of
// the Ministry.
func TestEmploymentDeclarationsMatchTheOfficialSamples(t *testing.T) {
	for _, sample := range employmentSamples {
		t.Run(sample.variant, func(t *testing.T) {
			want := readEmploymentSample(t, sample.file)

			got, err := employmentPayload(sample.variant, want)
			if err != nil {
				t.Fatalf("the payload of %s could not be built: %v", sample.variant, err)
			}

			if diff := compareFields(got, want); len(diff) > 0 {
				t.Errorf("the payload of %s does not match the official sample:\n  %s",
					sample.variant, strings.Join(diff, "\n  "))
			}
		})
	}
}

// employmentPayload turns the fields of a reference payload into the payload
// the SDK produces, through the public model.
func employmentPayload(variant string, fields map[string]any) (map[string]any, error) {
	raw, err := json.Marshal(fields)
	if err != nil {
		return nil, err
	}

	if strings.HasPrefix(variant, "WebE3") {
		var payload wire.HirePayload
		if err := json.Unmarshal(raw, &payload); err != nil {
			return nil, err
		}
		declaration := hireDeclarationOf(payload)
		return hireFieldsOf(HireVariant(variant), declaration)
	}

	var payload wire.TerminationPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}
	declaration := terminationDeclarationOf(payload)

	return terminationFieldsOf(TerminationVariant(variant), declaration)
}

// readEmploymentSample reads a reference payload of the family and returns
// the fields of the declaration.
func readEmploymentSample(t *testing.T, name string) map[string]any {
	t.Helper()

	path := filepath.Join("testdata", "employment", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read %s: %v", path, err)
	}

	var document map[string]any
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatalf("Failed to decode %s: %v", path, err)
	}

	// The root holds the object of the declaration.
	for _, value := range document {
		object, ok := value.(map[string]any)
		if !ok {
			t.Fatalf("The root of %s does not hold an object", path)
		}
		for _, inner := range object {
			fields, ok := inner.(map[string]any)
			if !ok {
				t.Fatalf("The declaration of %s is not an object", path)
			}
			return fields
		}
	}

	t.Fatalf("The declaration of %s is empty", path)

	return nil
}

// compareFields reports the fields of two payloads that differ.
func compareFields(got, want map[string]any) []string {
	normalisedGot := normaliseNumbers(got)
	normalisedWant := normaliseNumbers(want)

	var differences []string

	for name, value := range normalisedWant {
		other, ok := normalisedGot[name]
		if !ok {
			differences = append(differences, fmt.Sprintf("%s is missing", name))
			continue
		}
		if isDocumentField(name) && !isValidBase64(value) {
			// One of the published samples carries a document that is not
			// valid Base64. The SDK cannot reproduce a document it is unable
			// to decode, so the field is left out of the comparison.
			continue
		}
		if !reflect.DeepEqual(other, value) {
			differences = append(differences, fmt.Sprintf("%s is %v, want %v", name, other, value))
		}
	}

	for name := range normalisedGot {
		if _, ok := normalisedWant[name]; !ok {
			differences = append(differences, fmt.Sprintf("%s should not be there", name))
		}
	}

	sort.Strings(differences)

	return differences
}

// documentFields are the fields that carry a document.
var documentFields = map[string]bool{
	"f_file":         true,
	"f_foreign_file": true,
	"f_young_file":   true,
	"f_file_symbash": true,
}

// isDocumentField reports whether a field carries a document.
func isDocumentField(name string) bool {
	return documentFields[name]
}

// isValidBase64 reports whether a value is valid Base64.
func isValidBase64(value any) bool {
	text, ok := value.(string)
	if !ok {
		return true
	}

	_, err := base64.StdEncoding.DecodeString(text)

	return err == nil
}

// normaliseNumbers rewrites the numeric fields into a canonical form, because
// the samples spell them differently: "038,0" and "40,0" are both valid.
func normaliseNumbers(fields map[string]any) map[string]any {
	for name, digits := range numericFields {
		value, ok := fields[name]
		if !ok {
			continue
		}
		text, ok := value.(string)
		if !ok {
			continue
		}
		normalised, err := normaliseNumber(text, digits)
		if err != nil {
			continue
		}
		fields[name] = normalised
	}

	return fields
}

// normaliseNumber rewrites a number of the API into its canonical form.
func normaliseNumber(text string, digits int) (string, error) {
	plain := strings.ReplaceAll(text, ".", "")
	plain = strings.ReplaceAll(plain, ",", ".")

	value, err := strconv.ParseFloat(plain, 64)
	if err != nil {
		return "", err
	}

	return strconv.FormatFloat(value, 'f', digits, 64), nil
}
