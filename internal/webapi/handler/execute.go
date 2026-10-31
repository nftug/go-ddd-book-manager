package handler

import (
	"context"
	"net/http"

	dc "go-ddd-book-manager/internal/domain/common"
	"go-ddd-book-manager/internal/webapi/auth"

	"github.com/labstack/echo/v4"
)

func JSON[T any](
	c echo.Context,
	status int,
	execute func(context.Context, dc.Actor) (T, error),
) error {
	actor, err := auth.ActorFrom(c)
	if err != nil {
		return err
	}
	return jsonResult(c, status, func() (T, error) {
		return execute(c.Request().Context(), actor)
	})
}

func IdentityJSON[T any](
	c echo.Context,
	status int,
	execute func(context.Context, auth.Identity) (T, error),
) error {
	identity, err := auth.IdentityFrom(c)
	if err != nil {
		return err
	}
	return jsonResult(c, status, func() (T, error) {
		return execute(c.Request().Context(), identity)
	})
}

func jsonResult[T any](c echo.Context, status int, execute func() (T, error)) error {
	result, err := execute()
	if err != nil {
		return err
	}
	return c.JSON(status, result)
}

func NoContent(c echo.Context, execute func(context.Context, dc.Actor) error) error {
	actor, err := auth.ActorFrom(c)
	if err != nil {
		return err
	}
	if err = execute(c.Request().Context(), actor); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}
