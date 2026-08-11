// Package errs provides types and support related to web error functionality.
package errs

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Error represents an error in the system with an associated HTTP status.
type Error struct {
	Code    int    `json:"-"`
	Message string `json:"message"`
}

// New constructs an error based on an app error.
func New(code int, err error) *Error {
	return &Error{
		Code:    code,
		Message: err.Error(),
	}
}

// Newf constructs an error based on a error message.
func Newf(code int, format string, v ...any) *Error {
	return &Error{
		Code:    code,
		Message: fmt.Sprintf(format, v...),
	}
}

// Error implements the error interface.
func (e *Error) Error() string {
	return e.Message
}

// HTTPStatus implements the web package Statuser interface so the web framework
// can use the correct http status.
func (e *Error) HTTPStatus() int {
	return e.Code
}

// FieldError is used to indicate an error with a specific request field.
type FieldError struct {
	Field string `json:"field"`
	Err   string `json:"error"`
}

// FieldErrors represents a collection of field errors.
type FieldErrors []FieldError

// NewFieldErrors creates a field errors collection with a single entry.
func NewFieldErrors(field string, err error) FieldErrors {
	return FieldErrors{
		{
			Field: field,
			Err:   err.Error(),
		},
	}
}

// Error implements the error interface.
func (fe FieldErrors) Error() string {
	d, err := json.Marshal(fe)
	if err != nil {
		return err.Error()
	}

	return string(d)
}

// HTTPStatus implements the web package Statuser interface.
func (fe FieldErrors) HTTPStatus() int {
	return http.StatusBadRequest
}
