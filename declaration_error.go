package ergani

import (
	"errors"
	"fmt"
)

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
