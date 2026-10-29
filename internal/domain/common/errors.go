package common

import (
	"errors"
	"fmt"
)

var (
	ErrForbidden = errors.New("forbidden")
)

type ValidationError struct {
	Field string
	Code  string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Code)
}
