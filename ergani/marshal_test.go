package ergani

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

// assertNoPathSyntax walks a decoded JSON document and fails if any object key
// contains the ">" path separator. The standard library's encoding/json does
// not understand that syntax, so it used to leak into the payload as a literal
// key name ("Details>CardDetails") instead of a nested object.
func assertNoPathSyntax(t *testing.T, path string, value interface{}) {
	t.Helper()

	switch v := value.(type) {
	case map[string]interface{}:
		for key, child := range v {
			if strings.Contains(key, ">") {
				t.Errorf("%s: key %q must not use the '>' path syntax", path, key)
			}
			assertNoPathSyntax(t, path+"."+key, child)
		}
	case []interface{}:
		for i, child := range v {
			assertNoPathSyntax(t, path+"["+strconv.Itoa(i)+"]", child)
		}
	}
}

// decodeObject marshals v and decodes it into a generic object.
func decodeObject(t *testing.T, v interface{}) map[string]interface{} {
	t.Helper()

	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("Failed to marshal value: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal value: %v", err)
	}

	assertNoPathSyntax(t, "payload", decoded)
	return decoded
}

// nestedObject returns the object stored at the given path.
func nestedObject(t *testing.T, obj map[string]interface{}, path ...string) map[string]interface{} {
	t.Helper()

	current := obj
	for _, key := range path {
		child, ok := current[key].(map[string]interface{})
		if !ok {
			t.Fatalf("Expected %q to be a nested object, got %T", key, current[key])
		}
		current = child
	}
	return current
}

func TestCompanyWorkCard_MarshalJSON_NestsDetails(t *testing.T) {
	card := CompanyWorkCard{
		EmployerTaxID:        "999999999",
		BusinessBranchNumber: 1,
		Comments:             "night shift",
		CardDetails: []WorkCard{
			{
				EmployeeTaxID:            "123456789",
				EmployeeLastName:         "Doe",
				EmployeeFirstName:        "John",
				WorkCardMovementType:     Arrival,
				WorkCardSubmissionDate:   Date{Time: time.Date(2025, 7, 10, 0, 0, 0, 0, time.UTC)},
				WorkCardMovementDateTime: DateTime{Time: time.Date(2025, 7, 10, 9, 0, 0, 0, time.UTC)},
			},
		},
	}

	decoded := decodeObject(t, card)

	if decoded["f_afm_ergodoti"] != "999999999" {
		t.Errorf("Expected f_afm_ergodoti to be '999999999', got %v", decoded["f_afm_ergodoti"])
	}
	if decoded["f_aa"] != float64(1) {
		t.Errorf("Expected f_aa to be 1, got %v", decoded["f_aa"])
	}
	if decoded["f_comments"] != "night shift" {
		t.Errorf("Expected f_comments to be 'night shift', got %v", decoded["f_comments"])
	}

	details := nestedObject(t, decoded, "Details")
	cards, ok := details["CardDetails"].([]interface{})
	if !ok {
		t.Fatalf("Expected Details.CardDetails to be an array, got %T", details["CardDetails"])
	}
	if len(cards) != 1 {
		t.Fatalf("Expected 1 work card, got %d", len(cards))
	}

	entry, ok := cards[0].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected work card to be an object, got %T", cards[0])
	}
	if entry["f_type"] != ArrivalCode {
		t.Errorf("Expected f_type to be %q, got %v", ArrivalCode, entry["f_type"])
	}
	if entry["f_afm"] != "123456789" {
		t.Errorf("Expected f_afm to be '123456789', got %v", entry["f_afm"])
	}
	if entry["f_reference_date"] != "10/07/2025" {
		t.Errorf("Expected f_reference_date to be '10/07/2025', got %v", entry["f_reference_date"])
	}
}

func TestCompanyWorkCard_MarshalJSON_EmptyDetails(t *testing.T) {
	decoded := decodeObject(t, CompanyWorkCard{EmployerTaxID: "999999999"})

	details := nestedObject(t, decoded, "Details")
	cards, ok := details["CardDetails"].([]interface{})
	if !ok {
		t.Fatalf("Expected Details.CardDetails to be an array, got %T", details["CardDetails"])
	}
	if len(cards) != 0 {
		t.Errorf("Expected an empty CardDetails array, got %d entries", len(cards))
	}

	if _, ok := decoded["f_comments"]; ok {
		t.Error("Expected f_comments to be omitted when empty")
	}
}

func TestCompanyOvertime_MarshalJSON_NestsEmployees(t *testing.T) {
	overtime := CompanyOvertime{
		BusinessBranchNumber: 1,
		SEPEServiceCode:      "12345",
		EmployeeOvertimes: []Overtime{
			{
				EmployeeTaxID:         "123456789",
				EmployeeSSN:           "987654321",
				OvertimeJustification: ExceptionalWorkload,
				OvertimeStartTime:     Time{Time: time.Date(2025, 7, 10, 18, 0, 0, 0, time.UTC)},
				OvertimeEndTime:       Time{Time: time.Date(2025, 7, 10, 20, 0, 0, 0, time.UTC)},
				OvertimeDate:          Date{Time: time.Date(2025, 7, 10, 0, 0, 0, 0, time.UTC)},
				OvertimeCancellation:  Bool(false),
				WeeklyWorkdaysNumber:  5,
			},
		},
	}

	decoded := decodeObject(t, overtime)

	if decoded["f_aa_pararthmatos"] != float64(1) {
		t.Errorf("Expected f_aa_pararthmatos to be 1, got %v", decoded["f_aa_pararthmatos"])
	}

	employees := nestedObject(t, decoded, "Ergazomenoi")
	entries, ok := employees["OvertimeErgazomenosDate"].([]interface{})
	if !ok {
		t.Fatalf("Expected Ergazomenoi.OvertimeErgazomenosDate to be an array, got %T", employees["OvertimeErgazomenosDate"])
	}
	if len(entries) != 1 {
		t.Fatalf("Expected 1 overtime entry, got %d", len(entries))
	}

	entry, ok := entries[0].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected overtime entry to be an object, got %T", entries[0])
	}
	if entry["f_reason"] != ExceptionalWorkloadCode {
		t.Errorf("Expected f_reason to be %q, got %v", ExceptionalWorkloadCode, entry["f_reason"])
	}
	if entry["f_cancellation"] != "0" {
		t.Errorf("Expected f_cancellation to be \"0\", got %v", entry["f_cancellation"])
	}
	if entry["f_from"] != "18:00" {
		t.Errorf("Expected f_from to be '18:00', got %v", entry["f_from"])
	}
}

func TestEmployeeDailySchedule_MarshalJSON_NestsAnalytics(t *testing.T) {
	schedule := EmployeeDailySchedule{
		EmployeeTaxID:     "123456789",
		EmployeeLastName:  "Doe",
		EmployeeFirstName: "John",
		ScheduleDate:      Date{Time: time.Date(2025, 7, 10, 0, 0, 0, 0, time.UTC)},
		WorkdayDetails: []WorkdayDetails{
			{
				WorkType:  WorkFromOffice,
				StartTime: Time{Time: time.Date(2025, 7, 10, 9, 0, 0, 0, time.UTC)},
				EndTime:   Time{Time: time.Date(2025, 7, 10, 17, 0, 0, 0, time.UTC)},
			},
		},
	}

	decoded := decodeObject(t, schedule)

	if decoded["f_date"] != "10/07/2025" {
		t.Errorf("Expected f_date to be '10/07/2025', got %v", decoded["f_date"])
	}

	analytics := nestedObject(t, decoded, "ErgazomenosAnalytics")
	details, ok := analytics["ErgazomenosWTOAnalytics"].([]interface{})
	if !ok {
		t.Fatalf("Expected ErgazomenosAnalytics.ErgazomenosWTOAnalytics to be an array, got %T", analytics["ErgazomenosWTOAnalytics"])
	}
	if len(details) != 1 {
		t.Fatalf("Expected 1 workday detail, got %d", len(details))
	}

	detail, ok := details[0].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected workday detail to be an object, got %T", details[0])
	}
	if detail["f_type"] != WorkFromOfficeCode {
		t.Errorf("Expected f_type to be %q, got %v", WorkFromOfficeCode, detail["f_type"])
	}
	if detail["f_to"] != "17:00" {
		t.Errorf("Expected f_to to be '17:00', got %v", detail["f_to"])
	}
}

func TestCompanyDailySchedule_MarshalJSON_NestsEmployees(t *testing.T) {
	schedule := CompanyDailySchedule{
		BusinessBranchNumber: 1,
		EmployeeSchedules: []EmployeeDailySchedule{
			{
				EmployeeTaxID: "123456789",
				ScheduleDate:  Date{Time: time.Date(2025, 7, 10, 0, 0, 0, 0, time.UTC)},
				WorkdayDetails: []WorkdayDetails{
					{WorkType: WorkFromHome},
				},
			},
		},
	}

	decoded := decodeObject(t, schedule)

	employees := nestedObject(t, decoded, "Ergazomenoi")
	entries, ok := employees["ErgazomenoiWTO"].([]interface{})
	if !ok {
		t.Fatalf("Expected Ergazomenoi.ErgazomenoiWTO to be an array, got %T", employees["ErgazomenoiWTO"])
	}
	if len(entries) != 1 {
		t.Fatalf("Expected 1 employee schedule, got %d", len(entries))
	}

	entry, ok := entries[0].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected employee schedule to be an object, got %T", entries[0])
	}

	analytics := nestedObject(t, entry, "ErgazomenosAnalytics")
	details, ok := analytics["ErgazomenosWTOAnalytics"].([]interface{})
	if !ok {
		t.Fatalf("Expected ErgazomenosAnalytics.ErgazomenosWTOAnalytics to be an array, got %T", analytics["ErgazomenosWTOAnalytics"])
	}
	if len(details) != 1 {
		t.Fatalf("Expected 1 workday detail, got %d", len(details))
	}

	detail, ok := details[0].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected workday detail to be an object, got %T", details[0])
	}
	if detail["f_type"] != WorkFromHomeCode {
		t.Errorf("Expected f_type to be %q, got %v", WorkFromHomeCode, detail["f_type"])
	}
}

func TestEmployeeWeeklySchedule_MarshalJSON_NestsAnalytics(t *testing.T) {
	schedule := EmployeeWeeklySchedule{
		EmployeeTaxID: "123456789",
		ScheduleDay:   Weekday{Weekday: time.Monday},
		WorkdayDetails: []WorkdayDetails{
			{WorkType: WorkFromOffice},
		},
	}

	decoded := decodeObject(t, schedule)

	if decoded["f_day"] != float64(time.Monday) {
		t.Errorf("Expected f_day to be %d, got %v", time.Monday, decoded["f_day"])
	}

	analytics := nestedObject(t, decoded, "ErgazomenosAnalytics")
	details, ok := analytics["ErgazomenosWTOAnalytics"].([]interface{})
	if !ok {
		t.Fatalf("Expected ErgazomenosAnalytics.ErgazomenosWTOAnalytics to be an array, got %T", analytics["ErgazomenosWTOAnalytics"])
	}
	if len(details) != 1 {
		t.Fatalf("Expected 1 workday detail, got %d", len(details))
	}
}

func TestCompanyWeeklySchedule_MarshalJSON_NestsEmployees(t *testing.T) {
	schedule := CompanyWeeklySchedule{
		BusinessBranchNumber: 1,
		StartDate:            Date{Time: time.Date(2025, 7, 7, 0, 0, 0, 0, time.UTC)},
		EndDate:              Date{Time: time.Date(2025, 7, 13, 0, 0, 0, 0, time.UTC)},
		EmployeeSchedules: []EmployeeWeeklySchedule{
			{
				EmployeeTaxID: "123456789",
				ScheduleDay:   Weekday{Weekday: time.Tuesday},
				WorkdayDetails: []WorkdayDetails{
					{WorkType: RestDay},
				},
			},
		},
	}

	decoded := decodeObject(t, schedule)

	if decoded["f_from_date"] != "07/07/2025" {
		t.Errorf("Expected f_from_date to be '07/07/2025', got %v", decoded["f_from_date"])
	}
	if decoded["f_to_date"] != "13/07/2025" {
		t.Errorf("Expected f_to_date to be '13/07/2025', got %v", decoded["f_to_date"])
	}

	employees := nestedObject(t, decoded, "Ergazomenoi")
	entries, ok := employees["ErgazomenoiWTO"].([]interface{})
	if !ok {
		t.Fatalf("Expected Ergazomenoi.ErgazomenoiWTO to be an array, got %T", employees["ErgazomenoiWTO"])
	}
	if len(entries) != 1 {
		t.Fatalf("Expected 1 employee schedule, got %d", len(entries))
	}

	entry, ok := entries[0].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected employee schedule to be an object, got %T", entries[0])
	}

	analytics := nestedObject(t, entry, "ErgazomenosAnalytics")
	if _, ok := analytics["ErgazomenosWTOAnalytics"].([]interface{}); !ok {
		t.Fatalf("Expected ErgazomenosAnalytics.ErgazomenosWTOAnalytics to be an array, got %T", analytics["ErgazomenosWTOAnalytics"])
	}
}

// The full payload sent to the API must be nested as well: the submission
// wrapper is built with maps, and the entries inside it keep their nesting.
func TestSubmitWorkCard_SendsNestedPayload(t *testing.T) {
	var received map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/Authentication" {
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte(`{"accessToken": "test-token"}`)); err != nil {
				t.Errorf("Failed to write auth response: %v", err)
			}
			return
		}

		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`[]`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client, err := NewClient("testuser", "testpass", server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	_, err = client.SubmitWorkCard(context.Background(), []CompanyWorkCard{
		{
			EmployerTaxID:        "999999999",
			BusinessBranchNumber: 1,
			CardDetails: []WorkCard{
				{
					EmployeeTaxID:            "123456789",
					WorkCardMovementType:     Departure,
					WorkCardSubmissionDate:   Date{Time: time.Date(2025, 7, 10, 0, 0, 0, 0, time.UTC)},
					WorkCardMovementDateTime: DateTime{Time: time.Date(2025, 7, 10, 17, 0, 0, 0, time.UTC)},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("SubmitWorkCard failed: %v", err)
	}

	assertNoPathSyntax(t, "payload", received)

	cards := nestedObject(t, received, "Cards")
	entries, ok := cards["Card"].([]interface{})
	if !ok {
		t.Fatalf("Expected Cards.Card to be an array, got %T", cards["Card"])
	}
	if len(entries) != 1 {
		t.Fatalf("Expected 1 company work card, got %d", len(entries))
	}

	entry, ok := entries[0].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected company work card to be an object, got %T", entries[0])
	}

	details := nestedObject(t, entry, "Details")
	cardEntries, ok := details["CardDetails"].([]interface{})
	if !ok {
		t.Fatalf("Expected Details.CardDetails to be an array, got %T", details["CardDetails"])
	}
	if len(cardEntries) != 1 {
		t.Fatalf("Expected 1 work card, got %d", len(cardEntries))
	}

	card, ok := cardEntries[0].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected work card to be an object, got %T", cardEntries[0])
	}
	if card["f_type"] != DepartureCode {
		t.Errorf("Expected f_type to be %q, got %v", DepartureCode, card["f_type"])
	}
}
