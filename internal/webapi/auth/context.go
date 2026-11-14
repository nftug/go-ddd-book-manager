package auth

import (
	"errors"

	"github.com/labstack/echo/v5"
)

type contextKey int

const (
	identityKey contextKey = iota
)

func IdentityFrom(c *echo.Context) (Identity, error) {
	v, ok := c.Request().Context().Value(identityKey).(Identity)
	if !ok {
		return Identity{}, errors.New("missing verified identity")
	}
	return v, nil
}
