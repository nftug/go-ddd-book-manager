package common

import (
	"errors"
	"fmt"
)

var (
	ErrForbidden = errors.New("forbidden")
	ErrNotFound  = errors.New("not found")
	ErrConflict  = errors.New("conflict")
)

type ValidationError struct {
	field string
	code  string
}

func NewValidationError(field, code string) *ValidationError {
	return &ValidationError{field: field, code: code}
}

func (e *ValidationError) Field() string { return e.field }

func (e *ValidationError) Code() string { return e.code }

func (e *ValidationError) Error() string { return fmt.Sprintf("%s: %s", e.field, e.code) }
