package ergani

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func setupTestServer(t *testing.T) *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/Authentication", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST for auth, got %s", r.Method)
			return
		}
		var reqBody map[string]string
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatalf("Failed to decode auth request body: %v", err)
		}

		if reqBody["Username"] == "baduser" {
			w.WriteHeader(http.StatusUnauthorized)
			if _, err := w.Write([]byte(`{"message": "Invalid credentials"}`)); err != nil {
				t.Fatalf("Failed to write auth error response: %v", err)
			}
			return
		}

		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"accessToken": "test-token"}`)); err != nil {
			t.Fatalf("Failed to write auth success response: %v", err)
		}
	})

	mux.HandleFunc("/Documents/WRKCardSE", func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer test-token" {
			t.Errorf("Expected auth header 'Bearer test-token', got '%s'", authHeader)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body in mock server: %v", err)
		}
		if strings.Contains(string(bodyBytes), "FORCE_ERROR") {
			w.WriteHeader(http.StatusBadRequest)
			if _, err := w.Write([]byte(`{"msg": "Invalid data provided"}`)); err != nil {
				t.Fatalf("Failed to write error response for WRKCardSE: %v", err)
			}
			return
		}

		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`[{"id": "sub123", "protocol": "proto456", "submitDate": "10/07/2025 14:56"}]`)); err != nil {
			t.Fatalf("Failed to write success response for WRKCardSE: %v", err)
		}
	})

	mux.HandleFunc("/Documents/OvTime", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`[{"id": "sub456", "protocol": "proto789", "submitDate": "11/07/2025 10:00"}]`)); err != nil {
			t.Fatalf("Failed to write response for OvTime: %v", err)
		}
	})

	mux.HandleFunc("/Documents/WTODaily", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`[{"id": "sub789", "protocol": "proto123", "submitDate": "12/07/2025 11:00"}]`)); err != nil {
			t.Fatalf("Failed to write response for WTODaily: %v", err)
		}
	})

	mux.HandleFunc("/Documents/WTOWeek", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`[{"id": "sub111", "protocol": "proto222", "submitDate": "13/07/2025 12:00"}]`)); err != nil {
			t.Fatalf("Failed to write response for WTOWeek: %v", err)
		}
	})

	return httptest.NewServer(mux)
}

func TestNewClient_Success(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()

	client, err := NewClient("testuser", "testpass", server.URL)

	if err != nil {
		t.Fatalf("Expected no error during client creation, but got: %v", err)
	}
	if client == nil {
		t.Fatal("Expected client not to be nil")
	}
	if client.token != "" {
		t.Errorf("Expected token to be empty, but got '%s'", client.token)
	}
}

func TestNewClient_AuthFailure(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()

	client, err := NewClient("baduser", "badpass", server.URL)
	if err != nil {
		t.Fatalf("Expected no error during client creation, but got: %v", err)
	}

	_, err = client.SubmitWorkCard(context.Background(), []CompanyWorkCard{})

	if err == nil {
		t.Fatal("Expected an error during client creation for bad credentials, but got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("Expected error to be of type APIError, but got %T", err)
	}

	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status code 401, got %d", apiErr.StatusCode)
	}
}

func TestSubmitWorkCard_Success(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()

	client, _ := NewClient("testuser", "testpass", server.URL)

	workCards := []CompanyWorkCard{
		{
			EmployerTaxID:        "999999999",
			BusinessBranchNumber: 1,
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
		},
	}

	responses, err := client.SubmitWorkCard(context.Background(), workCards)
	if err != nil {
		t.Fatalf("Expected no error on SubmitWorkCard, but got: %v", err)
	}

	if len(responses) != 1 {
		t.Fatalf("Expected 1 submission response, got %d", len(responses))
	}

	expectedID := "sub123"
	if responses[0].ID != expectedID {
		t.Errorf("Expected submission ID '%s', got '%s'", expectedID, responses[0].ID)
	}
}

func TestSubmitWorkCard_APIError(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()

	client, _ := NewClient("testuser", "testpass", server.URL)

	workCards := []CompanyWorkCard{
		{Comments: "FORCE_ERROR"},
	}

	_, err := client.SubmitWorkCard(context.Background(), workCards)
	if err == nil {
		t.Fatal("Expected an APIError, but got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("Expected error to be of type APIError, but got %T", err)
	}
	if apiErr.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code 400, got %d", apiErr.StatusCode)
	}
}

func TestSubmitOvertime_Success(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()

	client, _ := NewClient("testuser", "testpass", server.URL)

	overtimes := []CompanyOvertime{
		{
			BusinessBranchNumber: 1,
			EmployeeOvertimes: []Overtime{
				{
					EmployeeTaxID:         "123456789",
					OvertimeJustification: ExceptionalWorkload,
				},
			},
		},
	}

	responses, err := client.SubmitOvertime(context.Background(), overtimes)
	if err != nil {
		t.Fatalf("Expected no error on SubmitOvertime, but got: %v", err)
	}

	if len(responses) != 1 {
		t.Fatalf("Expected 1 submission response, got %d", len(responses))
	}

	expectedID := "sub456"
	if responses[0].ID != expectedID {
		t.Errorf("Expected submission ID '%s', got '%s'", expectedID, responses[0].ID)
	}
}

func TestSubmitDailySchedule_Success(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()

	client, _ := NewClient("testuser", "testpass", server.URL)

	schedules := []CompanyDailySchedule{
		{
			BusinessBranchNumber: 1,
			EmployeeSchedules: []EmployeeDailySchedule{
				{
					EmployeeTaxID: "123456789",
				},
			},
		},
	}

	responses, err := client.SubmitDailySchedule(context.Background(), schedules)
	if err != nil {
		t.Fatalf("Expected no error on SubmitDailySchedule, but got: %v", err)
	}

	if len(responses) != 1 {
		t.Fatalf("Expected 1 submission response, got %d", len(responses))
	}

	expectedID := "sub789"
	if responses[0].ID != expectedID {
		t.Errorf("Expected submission ID '%s', got '%s'", expectedID, responses[0].ID)
	}
}

func TestSubmitWeeklySchedule_Success(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()

	client, _ := NewClient("testuser", "testpass", server.URL)

	schedules := []CompanyWeeklySchedule{
		{
			BusinessBranchNumber: 1,
			EmployeeSchedules: []EmployeeWeeklySchedule{
				{
					EmployeeTaxID: "123456789",
				},
			},
		},
	}

	responses, err := client.SubmitWeeklySchedule(context.Background(), schedules)
	if err != nil {
		t.Fatalf("Expected no error on SubmitWeeklySchedule, but got: %v", err)
	}

	if len(responses) != 1 {
		t.Fatalf("Expected 1 submission response, got %d", len(responses))
	}

	expectedID := "sub111"
	if responses[0].ID != expectedID {
		t.Errorf("Expected submission ID '%s', got '%s'", expectedID, responses[0].ID)
	}
}
