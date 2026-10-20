package ergani

import (
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestAPIError_Unwrap(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		want       error
	}{
		{"unauthorized", http.StatusUnauthorized, ErrUnauthorized},
		{"bad request", http.StatusBadRequest, ErrInvalidRequest},
		{"unprocessable entity", http.StatusUnprocessableEntity, ErrInvalidRequest},
		{"not found", http.StatusNotFound, ErrNotFound},
		{"rate limited", http.StatusTooManyRequests, ErrRateLimited},
		{"internal server error", http.StatusInternalServerError, ErrServer},
		{"bad gateway", http.StatusBadGateway, ErrServer},
		{"no content", http.StatusNoContent, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &APIError{StatusCode: tt.statusCode}

			if got := errors.Is(err, tt.want); got != (tt.want != nil) {
				t.Errorf("errors.Is(&APIError{StatusCode: %d}, %v) = %v, want %v",
					tt.statusCode, tt.want, got, tt.want != nil)
			}
		})
	}
}

func TestStatusCode(t *testing.T) {
	err := &APIError{StatusCode: http.StatusBadRequest, Message: "nope"}

	if got := StatusCode(err); got != http.StatusBadRequest {
		t.Errorf("StatusCode() = %d, want %d", got, http.StatusBadRequest)
	}
	if !IsStatusCode(err, http.StatusBadRequest) {
		t.Error("IsStatusCode() = false, want true")
	}
	if IsStatusCode(err, http.StatusNotFound) {
		t.Error("IsStatusCode() = true, want false")
	}
	if got := StatusCode(errors.New("other")); got != 0 {
		t.Errorf("StatusCode() = %d, want 0", got)
	}
}

func TestNewAPIError_Message(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{"object with a message", `{"message":"the AFM does not match"}`, "the AFM does not match"},
		{"object with a msg", `{"msg":"invalid data"}`, "invalid data"},
		{"object with a detail", `{"detail":"not allowed"}`, "not allowed"},
		{"bare string", `"the credentials are wrong"`, "the credentials are wrong"},
		{"html", `<html><body>Error</body></html>`, "<html><body>Error</body></html>"},
		{"empty", ``, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := newAPIError(http.StatusBadRequest, []byte(tt.body))

			if err.Message != tt.want {
				t.Errorf("newAPIError() message = %q, want %q", err.Message, tt.want)
			}
			if err.StatusCode != http.StatusBadRequest {
				t.Errorf("newAPIError() status = %d, want %d", err.StatusCode, http.StatusBadRequest)
			}
		})
	}
}

func TestParseSubmissions(t *testing.T) {
	tests := []struct {
		name string
		body string
		want int
	}{
		{"one submission", `[{"id":"92","protocol":"ΕΥΣ92","submitDate":"04/05/2022 01:13"}]`, 1},
		{"no submission", `[]`, 0},
		{"a date the API did not report", `[{"id":"92","protocol":"ΕΥΣ92"}]`, 1},
		{"a date the API did not understand", `[{"id":"92","protocol":"ΕΥΣ92","submitDate":"yesterday"}]`, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			submissions, err := parseSubmissions([]byte(tt.body))
			if err != nil {
				t.Fatalf("parseSubmissions() error = %v", err)
			}
			if len(submissions) != tt.want {
				t.Errorf("parseSubmissions() returned %d submissions, want %d", len(submissions), tt.want)
			}
		})
	}
}

func TestParseRetryAfter(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  int
	}{
		{"empty", "", 0},
		{"seconds", "3", 3},
		{"negative", "-1", 0},
		{"garbage", "later", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseRetryAfter(tt.value)
			if got != time.Duration(tt.want)*time.Second {
				t.Errorf("parseRetryAfter(%q) = %s, want %ds", tt.value, got, tt.want)
			}
		})
	}
}

// FuzzNewAPIError checks that the parser of the error bodies never panics,
// whatever the API answers.
func FuzzNewAPIError(f *testing.F) {
	f.Add([]byte(`{"message":"nope"}`))
	f.Add([]byte(`"bare"`))
	f.Add([]byte(`<html></html>`))
	f.Add([]byte(``))

	f.Fuzz(func(t *testing.T, body []byte) {
		err := newAPIError(http.StatusBadRequest, body)
		if err == nil {
			t.Fatal("newAPIError() returned nil")
		}
		if err.StatusCode != http.StatusBadRequest {
			t.Errorf("newAPIError() status = %d, want %d", err.StatusCode, http.StatusBadRequest)
		}
	})
}
