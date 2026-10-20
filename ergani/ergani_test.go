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

// setupTestServer creates a mock HTTP server to simulate the Ergani API.
func setupTestServer(t *testing.T) *httptest.Server {
	mux := http.NewServeMux()

	// Mock Authentication Endpoint
	mux.HandleFunc("/Authentication", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST for auth, got %s", r.Method)
			return
		}
		var reqBody map[string]string
		json.NewDecoder(r.Body).Decode(&reqBody)

		// Check for bad credentials for failure testing
		if reqBody["Username"] == "baduser" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"message": "Invalid credentials"}`))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"accessToken": "test-token"}`))
	})

	// Mock Document Submission Endpoint
	mux.HandleFunc("/Documents/WRKCardSE", func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer test-token" {
			t.Errorf("Expected auth header 'Bearer test-token', got '%s'", authHeader)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Simulate an API error based on a specific input
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body in mock server: %v", err)
		}
		if strings.Contains(string(bodyBytes), "FORCE_ERROR") {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"msg": "Invalid data provided"}`))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"id": "sub123", "protocol": "proto456", "submitDate": "10/07/2025 14:56"}]`))
	})

	return httptest.NewServer(mux)
}

func TestNewClient_Success(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()

	client, err := NewClient(context.Background(), "testuser", "testpass", server.URL)

	if err != nil {
		t.Fatalf("Expected no error during client creation, but got: %v", err)
	}
	if client == nil {
		t.Fatal("Expected client not to be nil")
	}
	if client.token != "test-token" {
		t.Errorf("Expected token 'test-token', got '%s'", client.token)
	}
}

func TestNewClient_AuthFailure(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()

	_, err := NewClient(context.Background(), "baduser", "badpass", server.URL)

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

	client, _ := NewClient(context.Background(), "testuser", "testpass", server.URL)

	workCards := []CompanyWorkCard{
		{
			EmployerTaxID:        "999999999",
			BusinessBranchNumber: 1,
			CardDetails: []WorkCard{
				{
					EmployeeTaxID:            "123456789",
					EmployeeLastName:         "Doe",
					EmployeeFirstName:        "John",
					WorkCardMovementType:     Arrival, // Add this required field
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

	client, _ := NewClient(context.Background(), "testuser", "testpass", server.URL)

	// Using a specific value to trigger the error in the mock server
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
