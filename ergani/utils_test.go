package ergani

import (
	"encoding/json"
	"testing"
	"time"
)

func TestMapWorkCardMovementType(t *testing.T) {
	// Test valid cases
	val, err := mapWorkCardMovementType(Arrival)
	if err != nil {
		t.Errorf("Unexpected error for Arrival: %v", err)
	}
	if val != "0" {
		t.Errorf("Expected Arrival to map to '0', got '%s'", val)
	}

	val, err = mapWorkCardMovementType(Departure)
	if err != nil {
		t.Errorf("Unexpected error for Departure: %v", err)
	}
	if val != "1" {
		t.Errorf("Expected Departure to map to '1', got '%s'", val)
	}

	// Test invalid case
	_, err = mapWorkCardMovementType(WorkCardMovementType("INVALID"))
	if err == nil {
		t.Error("Expected error for invalid WorkCardMovementType, got nil")
	}
}

func TestMapScheduleWorkType(t *testing.T) {
	// Test valid cases
	val, err := mapScheduleWorkType(WorkFromOffice)
	if err != nil {
		t.Errorf("Unexpected error for WorkFromOffice: %v", err)
	}
	if val != "ΕΡΓ" {
		t.Errorf("Expected WorkFromOffice to map to 'ΕΡΓ', got '%s'", val)
	}

	val, err = mapScheduleWorkType(WorkFromHome)
	if err != nil {
		t.Errorf("Unexpected error for WorkFromHome: %v", err)
	}
	if val != "ΤΗΛ" {
		t.Errorf("Expected WorkFromHome to map to 'ΤΗΛ', got '%s'", val)
	}

	// Test invalid case
	_, err = mapScheduleWorkType(ScheduleWorkType("INVALID"))
	if err == nil {
		t.Error("Expected error for invalid ScheduleWorkType, got nil")
	}
}

func TestCustomTimeTypes_MarshalJSON(t *testing.T) {
	// Test Time
	tm := Time{Time: time.Date(0, 1, 1, 14, 30, 0, 0, time.UTC)}
	b, err := json.Marshal(tm)
	if err != nil {
		t.Fatalf("Time MarshalJSON failed: %v", err)
	}
	if string(b) != `"14:30"` {
		t.Errorf(`Expected Time to marshal to "14:30", got %s`, string(b))
	}

	// Test Date
	dt := Date{Time: time.Date(2025, 7, 10, 0, 0, 0, 0, time.UTC)}
	b, err = json.Marshal(dt)
	if err != nil {
		t.Fatalf("Date MarshalJSON failed: %v", err)
	}
	if string(b) != `"10/07/2025"` {
		t.Errorf(`Expected Date to marshal to "10/07/2025", got %s`, string(b))
	}

	// Test Bool
	bTrue := Bool(true)
	b, err = json.Marshal(bTrue)
	if err != nil {
		t.Fatalf("Bool(true) MarshalJSON failed: %v", err)
	}
	if string(b) != `"1"` {
		t.Errorf(`Expected Bool(true) to marshal to "1", got %s`, string(b))
	}

	bFalse := Bool(false)
	b, err = json.Marshal(bFalse)
	if err != nil {
		t.Fatalf("Bool(false) MarshalJSON failed: %v", err)
	}
	if string(b) != `"0"` {
		t.Errorf(`Expected Bool(false) to marshal to "0", got %s`, string(b))
	}
}

// Add comprehensive table-driven tests for all enum mappings
func TestEnumMappings(t *testing.T) {
	t.Run("WorkCardMovementType", func(t *testing.T) {
		tests := []struct {
			name     string
			input    WorkCardMovementType
			expected string
			hasError bool
		}{
			{"Arrival", Arrival, "0", false},
			{"Departure", Departure, "1", false},
			{"Invalid", WorkCardMovementType("INVALID"), "", true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := mapWorkCardMovementType(tt.input)
				if tt.hasError {
					if err == nil {
						t.Errorf("Expected error for input %v, got nil", tt.input)
					}
				} else {
					if err != nil {
						t.Errorf("Unexpected error for input %v: %v", tt.input, err)
					}
					if result != tt.expected {
						t.Errorf("Expected %s, got %s", tt.expected, result)
					}
				}
			})
		}
	})

	t.Run("LateDeclarationJustificationType", func(t *testing.T) {
		tests := []struct {
			name     string
			input    LateDeclarationJustificationType
			expected string
			hasError bool
		}{
			{"PowerOutage", PowerOutage, "001", false},
			{"EmployerSystemsUnavailable", EmployerSystemsUnavailable, "002", false},
			{"ErganiSystemsUnavailable", ErganiSystemsUnavailable, "003", false},
			{"Invalid", LateDeclarationJustificationType("INVALID"), "", true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := mapLateDeclarationJustification(tt.input)
				if tt.hasError {
					if err == nil {
						t.Errorf("Expected error for input %v, got nil", tt.input)
					}
				} else {
					if err != nil {
						t.Errorf("Unexpected error for input %v: %v", tt.input, err)
					}
					if result != tt.expected {
						t.Errorf("Expected %s, got %s", tt.expected, result)
					}
				}
			})
		}
	})

	t.Run("ScheduleWorkType", func(t *testing.T) {
		tests := []struct {
			name     string
			input    ScheduleWorkType
			expected string
			hasError bool
		}{
			{"WorkFromOffice", WorkFromOffice, "ΕΡΓ", false},
			{"WorkFromHome", WorkFromHome, "ΤΗΛ", false},
			{"RestDay", RestDay, "ΑΝ", false},
			{"Absent", Absent, "ΜΕ", false},
			{"Invalid", ScheduleWorkType("INVALID"), "", true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := mapScheduleWorkType(tt.input)
				if tt.hasError {
					if err == nil {
						t.Errorf("Expected error for input %v, got nil", tt.input)
					}
				} else {
					if err != nil {
						t.Errorf("Unexpected error for input %v: %v", tt.input, err)
					}
					if result != tt.expected {
						t.Errorf("Expected %s, got %s", tt.expected, result)
					}
				}
			})
		}
	})
}
