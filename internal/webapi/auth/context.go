package auth

import (
	"errors"
	dc "go-ddd-book-manager/internal/domain/common"

	"github.com/labstack/echo/v5"
)

type contextKey int

const (
	identityKey contextKey = iota
	actorKey
)

func IdentityFrom(c *echo.Context) (Identity, error) {
	v, ok := c.Request().Context().Value(identityKey).(Identity)
	if !ok {
		return Identity{}, errors.New("missing verified identity")
	}
	return v, nil
}

func ActorFrom(c *echo.Context) (dc.Actor, error) {
	v, ok := c.Request().Context().Value(actorKey).(dc.Actor)
	if !ok {
		return dc.Actor{}, errors.New("missing actor")
	}
	return v, nil
}
