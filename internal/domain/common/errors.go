package common

import (
	"errors"
)

var (
	ErrForbidden = errors.New("forbidden")
	ErrNotFound  = errors.New("not found")
)

type ValidationError struct {
	code string
}

func NewValidationError(code string) *ValidationError {
	return &ValidationError{code: code}
}

func (e *ValidationError) Code() string { return e.code }

func (e *ValidationError) Error() string { return e.code }
