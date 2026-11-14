package handler

import (
	"context"
	"errors"
	"net/http"

	uc "go-ddd-book-manager/internal/application/user/command"
	ud "go-ddd-book-manager/internal/application/user/dto"
	dc "go-ddd-book-manager/internal/domain/common"
	"go-ddd-book-manager/internal/webapi/auth"

	"github.com/labstack/echo/v5"
	"github.com/samber/do/v2"
)

type Executor struct{ syncUser uc.SyncUser }

func NewExecutor(i do.Injector) (*Executor, error) {
	return &Executor{syncUser: do.MustInvoke[uc.SyncUser](i)}, nil
}

func (e *Executor) OK(
	c *echo.Context,
	execute func(context.Context, dc.Actor) (any, error),
) error {
	return e.json(c, http.StatusOK, execute)
}

func (e *Executor) Created(
	c *echo.Context,
	execute func(context.Context, dc.Actor) (any, error),
) error {
	return e.json(c, http.StatusCreated, execute)
}

func (e *Executor) NoContent(
	c *echo.Context,
	execute func(context.Context, dc.Actor) error,
) error {
	if _, err := e.execute(c, func(ctx context.Context, actor dc.Actor) (any, error) {
		return nil, execute(ctx, actor)
	}); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func (e *Executor) json(
	c *echo.Context,
	status int,
	execute func(context.Context, dc.Actor) (any, error),
) error {
	result, err := e.execute(c, execute)
	if err != nil {
		return err
	}
	return c.JSON(status, result)
}

func (e *Executor) execute(
	c *echo.Context,
	execute func(context.Context, dc.Actor) (any, error),
) (any, error) {
	identity, err := auth.IdentityFrom(c)
	if err != nil {
		return dc.Actor{}, err
	}
	actor, err := e.syncUser.Execute(
		c.Request().Context(),
		ud.SyncUserRequest{
			ID:    identity.ID,
			Name:  identity.Name,
			Email: identity.Email,
			Role:  identity.Role,
		},
	)
	if _, ok := errors.AsType[*dc.ValidationError](err); ok {
		return dc.Actor{}, auth.ErrUnauthorized
	}
	if err != nil {
		return dc.Actor{}, err
	}
	return execute(c.Request().Context(), actor)
}
