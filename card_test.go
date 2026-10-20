package ergani

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

// The samples in testdata/official are the payloads the Ministry of Labour
// publishes for the API. The SDK must produce the same payload, field for
// field, when it is given the same declaration.
//
// This test lives in the package itself because it exercises the translation
// from the public models to the wire format, which is unexported.

// readOfficial reads a reference payload.
func readOfficial(t *testing.T, name string) map[string]any {
	t.Helper()

	path := filepath.Join("testdata", "official", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read %s: %v", path, err)
	}

	// The published work card sample carries a trailing comma, which is not
	// valid JSON. The commas that precede a closing brace are removed first.
	cleaned := removeTrailingCommas(string(data))

	var decoded map[string]any
	if err := json.Unmarshal([]byte(cleaned), &decoded); err != nil {
		t.Fatalf("Failed to decode %s: %v", path, err)
	}

	return decoded
}

// removeTrailingCommas removes the commas that precede a closing brace or a
// closing bracket.
func removeTrailingCommas(s string) string {
	out := make([]byte, 0, len(s))
	inString := false
	escaped := false

	for i := 0; i < len(s); i++ {
		c := s[i]

		if inString {
			out = append(out, c)
			switch {
			case escaped:
				escaped = false
			case c == '\\':
				escaped = true
			case c == '"':
				inString = false
			}
			continue
		}

		if c == '"' {
			inString = true
			out = append(out, c)
			continue
		}

		if c == ',' {
			j := i + 1
			for j < len(s) && (s[j] == ' ' || s[j] == '\t' || s[j] == '\n' || s[j] == '\r') {
				j++
			}
			if j < len(s) && (s[j] == '}' || s[j] == ']') {
				continue
			}
		}

		out = append(out, c)
	}

	return string(out)
}

// assertSameJSON compares two decoded JSON documents.
func assertSameJSON(t *testing.T, got, want map[string]any, context string) {
	t.Helper()

	if !reflect.DeepEqual(got, want) {
		gotJSON, _ := json.MarshalIndent(got, "", "  ")
		wantJSON, _ := json.MarshalIndent(want, "", "  ")
		t.Errorf("%s: the payload does not match the official sample\n got: %s\nwant: %s",
			context, gotJSON, wantJSON)
	}
}

// decodePayload marshals a payload and decodes it into a generic document.
func decodePayload(t *testing.T, payload any) map[string]any {
	t.Helper()

	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal the payload: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("Failed to decode the payload: %v", err)
	}

	return decoded
}

// TestWorkCardMatchesOfficialSample checks the payload of a work card
// against the sample the Ministry publishes.
func TestWorkCardMatchesOfficialSample(t *testing.T) {
	payload, err := buildWorkCards([]WorkCardDeclaration{
		{
			EmployerTaxID: "111111111",
			BranchNumber:  "0",
			Comments:      "test",
			Cards: []WorkCard{
				{
					EmployeeTaxID:     "999999999",
					EmployeeLastName:  "ΤΕΣΤ",
					EmployeeFirstName: "ΤΕΣΤ",
					MovementType:      MovementDeparture,
					ReferenceDate:     time.Date(2022, 4, 21, 0, 0, 0, 0, time.UTC),
					MovementDateTime:  time.Date(2022, 4, 21, 12, 45, 0, 0, time.UTC),
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("buildWorkCards() error = %v", err)
	}

	got := decodePayload(t, payload)
	want := readOfficial(t, "Card.json")

	// The sample carries the moment of the movement without an offset, while
	// the SDK always sends one. The offset is added to the expected value.
	movement := want["Cards"].(map[string]any)["Card"].([]any)[0].(map[string]any)["Details"].(map[string]any)["CardDetails"].([]any)[0].(map[string]any)
	movement["f_date"] = "2022-04-21T12:45:00Z"

	// The published sample file uses an empty string for the justification
	// of a late declaration, while the request example of the official
	// interoperability guide uses null. The schema allows a code or spaces,
	// so the SDK sends null, the way the guide does.
	if movement["f_aitiologia"] == "" {
		movement["f_aitiologia"] = nil
	}

	assertSameJSON(t, got, want, "work card")
}

// TestDailyScheduleMatchesOfficialSample checks the payload of a daily
// schedule against the sample the Ministry publishes.
func TestDailyScheduleMatchesOfficialSample(t *testing.T) {
	payload, err := buildDailySchedules([]DailyScheduleDeclaration{
		{
			BranchNumber: "0",
			Comments:     "test",
			Schedules: []EmployeeDailySchedule{
				{
					EmployeeTaxID:     "999999999",
					EmployeeLastName:  "ΤΕΣΤ",
					EmployeeFirstName: "ΤΕΣΤ",
					Date:              time.Date(2022, 4, 11, 0, 0, 0, 0, time.UTC),
					Workday: []Workday{
						{
							WorkType:  WorkTypeOffice,
							StartTime: time.Date(2022, 4, 11, 9, 0, 0, 0, time.UTC),
							EndTime:   time.Date(2022, 4, 11, 18, 0, 0, 0, time.UTC),
						},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("buildDailySchedules() error = %v", err)
	}

	got := decodePayload(t, payload)
	want := readOfficial(t, "WTODaily.json")

	assertSameJSON(t, got, want, "daily schedule")
}

// TestWeeklyScheduleMatchesOfficialSample checks the payload of a weekly
// schedule against the sample the Ministry publishes.
func TestWeeklyScheduleMatchesOfficialSample(t *testing.T) {
	payload, err := buildWeeklySchedules([]WeeklyScheduleDeclaration{
		{
			BranchNumber: "0",
			Comments:     "test",
			StartDate:    time.Date(2022, 4, 26, 0, 0, 0, 0, time.UTC),
			EndDate:      time.Date(2022, 4, 26, 0, 0, 0, 0, time.UTC),
			Schedules: []EmployeeWeeklySchedule{
				{
					EmployeeTaxID:     "999999999",
					EmployeeLastName:  "ΤΕΣΤ",
					EmployeeFirstName: "ΤΕΣΤ",
					Day:               time.Tuesday,
					Workday: []Workday{
						{
							WorkType:  WorkTypeOffice,
							StartTime: time.Date(2022, 4, 26, 6, 0, 0, 0, time.UTC),
							EndTime:   time.Date(2022, 4, 26, 14, 0, 0, 0, time.UTC),
						},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("buildWeeklySchedules() error = %v", err)
	}

	got := decodePayload(t, payload)
	want := readOfficial(t, "WTOWEEKLY.json")

	assertSameJSON(t, got, want, "weekly schedule")
}
