package auth

import "errors"

var ErrUnauthorized = errors.New("unauthorized")
var ErrProvider = errors.New("identity provider unavailable")
