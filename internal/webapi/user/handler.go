package user

import (
	"context"
	uq "go-ddd-book-manager/internal/application/user/query"
	dc "go-ddd-book-manager/internal/domain/common"
	"go-ddd-book-manager/internal/webapi/auth"
	"go-ddd-book-manager/internal/webapi/handler"

	"github.com/labstack/echo/v5"
	"github.com/samber/do/v2"
)

type Handler struct {
	verifier    auth.TokenVerifier
	executor    *handler.Executor
	userDetails uq.GetUserDetails
}

func NewHandler(i do.Injector) (*Handler, error) {
	return &Handler{
		verifier:    do.MustInvoke[auth.TokenVerifier](i),
		executor:    do.MustInvoke[*handler.Executor](i),
		userDetails: do.MustInvoke[uq.GetUserDetails](i),
	}, nil
}

func (h *Handler) RegisterRoutes(e *echo.Echo) {
	group := e.Group("/users", auth.RequireIdentity(h.verifier))
	group.GET("/me", h.details)
}

// details godoc
// @ID GetCurrentUser
// @Summary Get the current user
// @Tags Users
// @Produce json
// @Success 200 {object} dto.UserResponse
// @Failure 401 {object} handler.ErrorResponse "Unauthorized"
// @Failure 404 {object} handler.ErrorResponse "Not Found"
// @Failure 500 {object} handler.ErrorResponse "Internal Server Error"
// @Header 401 {string} WWW-Authenticate "Bearer"
// @Security BearerAuth
// @Router /users/me [get]
func (h *Handler) details(c *echo.Context) error {
	return h.executor.OK(c, func(ctx context.Context, actor dc.Actor) (any, error) {
		return h.userDetails.Execute(ctx, actor)
	})
}
