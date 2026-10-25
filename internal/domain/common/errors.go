package common

import "errors"

var (
	ErrForbidden      = errors.New("forbidden")
	ErrInvalidRequest = errors.New("invalid request")
)
