package auth

import (
	"context"
	"errors"
	us "go-ddd-book-manager/internal/application/user/service"
	dc "go-ddd-book-manager/internal/domain/common"
	"strings"

	"github.com/labstack/echo/v5"
)

func RequireIdentity(v TokenVerifier) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			h := strings.Fields(c.Request().Header.Get("Authorization"))
			if len(h) != 2 || !strings.EqualFold(h[0], "Bearer") {
				return ErrUnauthorized
			}
			identity, err := v.Verify(c.Request().Context(), h[1])
			if err != nil {
				return err
			}

			ctx := context.WithValue(c.Request().Context(), identityKey, identity)
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}

func RequireUser(v TokenVerifier, r us.ResolveActor) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return RequireIdentity(v)(func(c *echo.Context) error {
			identity, err := IdentityFrom(c)
			if err != nil {
				return err
			}
			actor, err := r.Execute(c.Request().Context(), identity.ID)
			if errors.Is(err, dc.ErrNotFound) {
				return ErrUnauthorized
			}
			if err != nil {
				return err
			}

			ctx := context.WithValue(c.Request().Context(), actorKey, actor)
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		})
	}
}
