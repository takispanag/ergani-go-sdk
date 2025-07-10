package ergani

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// APIError represents a generic error returned by the Ergani API.
// It captures the HTTP status code, a parsed error message, and the full
// raw response body for debugging.
type APIError struct {
	StatusCode int
	Message    string
	Response   string
}

// newAPIError creates a new APIError from an http.Response. It reads the
// response body and attempts to parse a structured error message from it.
// If parsing fails, it falls back to using the raw body as the message.
func newAPIError(r *http.Response) *APIError {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return &APIError{
			StatusCode: r.StatusCode,
			Message:    "failed to read error response body",
		}
	}

	// The API can return error messages in different JSON fields ("message", "msg", "detail").
	// We try to find the most specific one.
	var errorResponse struct {
		Message string `json:"message"`
		Msg     string `json:"msg"`
		Detail  string `json:"detail"`
	}
	message := string(body) // Fallback message
	if json.Unmarshal(body, &errorResponse) == nil {
		if errorResponse.Message != "" {
			message = errorResponse.Message
		} else if errorResponse.Msg != "" {
			message = errorResponse.Msg
		} else if errorResponse.Detail != "" {
			message = errorResponse.Detail
		}
	}

	return &APIError{
		StatusCode: r.StatusCode,
		Message:    message,
		Response:   string(body),
	}
}

// Error implements the standard error interface.
func (e *APIError) Error() string {
	return fmt.Sprintf("API error (status %d): %s", e.StatusCode, e.Message)
}

// AuthenticationError is a specific type of error for authentication failures.
// It is used when authentication succeeds but the server doesn't return a token.
type AuthenticationError struct {
	Message string
}

// Error implements the standard error interface.
func (e *AuthenticationError) Error() string {
	return fmt.Sprintf("authentication failed: %s", e.Message)
}
