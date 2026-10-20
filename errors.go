package ergani

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// Sentinel errors that classify a failure. Use errors.Is to test for them.
//
// A failure reported by the API also carries an [*APIError] with the HTTP
// status code and the message of the service.
var (
	// ErrUnauthorized reports that the API rejected the credentials or the
	// access token.
	ErrUnauthorized = errors.New("ergani: unauthorized")

	// ErrInvalidRequest reports that the API rejected the payload.
	ErrInvalidRequest = errors.New("ergani: invalid request")

	// ErrNotFound reports that the API did not find the resource.
	ErrNotFound = errors.New("ergani: not found")

	// ErrRateLimited reports that the API asked the client to slow down.
	ErrRateLimited = errors.New("ergani: rate limited")

	// ErrServer reports a failure of the API itself.
	ErrServer = errors.New("ergani: server error")
)

// APIError is a failure reported by the Ergani API.
//
// Use errors.As to obtain one from an error, and errors.Is to classify it
// with one of the sentinel errors of this package.
type APIError struct {
	// StatusCode is the HTTP status code of the response.
	StatusCode int

	// Message is the message the API reported. It is the raw body when the
	// body was not a JSON object with a message.
	Message string

	// RawBody is the body of the response. It is empty when the body could
	// not be read.
	RawBody string
}

// Error implements the error interface.
func (e *APIError) Error() string {
	return fmt.Sprintf("ergani: api error (status %d): %s", e.StatusCode, e.Message)
}

// Unwrap returns the sentinel error that matches the status code of the
// response, or nil when the status code has no sentinel.
func (e *APIError) Unwrap() error {
	switch {
	case e.StatusCode == http.StatusUnauthorized:
		return ErrUnauthorized
	case e.StatusCode == http.StatusBadRequest, e.StatusCode == http.StatusUnprocessableEntity:
		return ErrInvalidRequest
	case e.StatusCode == http.StatusNotFound:
		return ErrNotFound
	case e.StatusCode == http.StatusTooManyRequests:
		return ErrRateLimited
	case e.StatusCode >= 500:
		return ErrServer
	default:
		return nil
	}
}

// ValidationError reports a declaration that the SDK rejected before it
// reached the API.
type ValidationError struct {
	// Field is the path of the offending field, for example
	// "declarations[0].cards[2].movementType".
	Field string

	// Value is the offending value.
	Value any

	// Reason explains what is wrong with the value.
	Reason string
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("ergani: invalid %s: %s", e.Field, e.Reason)
}

// AuthenticationError reports that the API did not return an access token
// after a successful authentication.
type AuthenticationError struct {
	// Reason explains why the token was missing.
	Reason string
}

// Error implements the error interface.
func (e *AuthenticationError) Error() string {
	return "ergani: authentication failed: " + e.Reason
}

// StatusCode returns the HTTP status code carried by err, or 0 when err does
// not carry one.
func StatusCode(err error) int {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode
	}

	return 0
}

// IsStatusCode reports whether err carries the given HTTP status code.
func IsStatusCode(err error, code int) bool {
	return StatusCode(err) == code
}

// newAPIError builds an APIError from a status code and the body of the
// response.
//
// The API reports its failures either as an object with a message, or as a
// bare JSON string. A body that is neither, such as the HTML error page of
// the web server, becomes the message itself.
func newAPIError(status int, body []byte) *APIError {
	apiErr := &APIError{StatusCode: status, RawBody: string(body)}

	var object struct {
		Message string `json:"message"`
		Msg     string `json:"msg"`
		Detail  string `json:"detail"`
	}
	if err := json.Unmarshal(body, &object); err == nil {
		switch {
		case object.Message != "":
			apiErr.Message = object.Message
			return apiErr
		case object.Msg != "":
			apiErr.Message = object.Msg
			return apiErr
		case object.Detail != "":
			apiErr.Message = object.Detail
			return apiErr
		}
	}

	var message string
	if err := json.Unmarshal(body, &message); err == nil && message != "" {
		apiErr.Message = message
		return apiErr
	}

	apiErr.Message = strings.TrimSpace(string(body))

	return apiErr
}

// DeclarationError is a failure of a submission.
//
// It carries the type of the declaration that was rejected, so that a
// caller can tell a rejected work card from a rejected overtime without
// looking at the status code.
type DeclarationError struct {
	// Code is the type of the declaration that was rejected.
	Code DocumentCode

	// Err is the failure the API reported.
	Err error
}

// Error implements the error interface.
func (e *DeclarationError) Error() string {
	return fmt.Sprintf("ergani: the declaration %s was rejected: %v", e.Code, e.Err)
}

// Unwrap returns the failure the API reported, so that errors.Is and
// errors.As keep working.
func (e *DeclarationError) Unwrap() error {
	return e.Err
}

// CodeOf returns the type of the declaration that a failure belongs to.
//
// It returns an empty code when the failure does not come from a
// declaration.
func CodeOf(err error) DocumentCode {
	var declarationErr *DeclarationError
	if errors.As(err, &declarationErr) {
		return declarationErr.Code
	}

	return ""
}
