package ergani

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestSubmissionResponse_UnmarshalJSON(t *testing.T) {
	jsonData := `{"id": "sub123", "protocol": "proto456", "submitDate": "10/07/2025 14:56"}`
	var resp SubmissionResponse

	err := json.Unmarshal([]byte(jsonData), &resp)
	if err != nil {
		t.Fatalf("Failed to unmarshal submission response: %v", err)
	}

	expectedDate, _ := time.Parse("02/01/2006 15:04", "10/07/2025 14:56")
	if !resp.SubmissionDate.Equal(expectedDate) {
		t.Errorf("Expected date %v, got %v", expectedDate, resp.SubmissionDate)
	}
	if resp.ID != "sub123" {
		t.Errorf("Expected ID 'sub123', got '%s'", resp.ID)
	}
}

func TestCompanyWorkCard_MarshalJSON(t *testing.T) {
	powerOutage := PowerOutage
	card := CompanyWorkCard{
		EmployerTaxID:        "999999999",
		BusinessBranchNumber: 1,
		CardDetails: []WorkCard{
			{
				EmployeeTaxID:                "123456789",
				WorkCardMovementType:         Arrival,
				LateDeclarationJustification: &powerOutage,
				WorkCardSubmissionDate:       Date{Time: time.Date(2025, 7, 10, 0, 0, 0, 0, time.UTC)},
				WorkCardMovementDateTime:     DateTime{Time: time.Date(2025, 7, 10, 14, 56, 0, 0, time.UTC)},
			},
		},
	}

	bytes, err := json.Marshal(card)
	if err != nil {
		t.Fatalf("Failed to marshal CompanyWorkCard: %v", err)
	}

	jsonString := string(bytes)

	if !strings.Contains(jsonString, `"f_afm_ergodoti":"999999999"`) {
		t.Error("Expected to find marshaled employer tax ID")
	}
	if !strings.Contains(jsonString, `"f_type":"0"`) {
		t.Error("Expected to find marshaled work card movement type '0'")
	}
	if !strings.Contains(jsonString, `"f_aitiologia":"001"`) {
		t.Error("Expected to find marshaled late declaration justification '001'")
	}
	if !strings.Contains(jsonString, `"f_reference_date":"10/07/2025"`) {
		t.Error("Expected to find correctly formatted date")
	}
}
