package user

import (
	"context"
	"errors"
	uc "go-ddd-book-manager/internal/application/user/command"
	ud "go-ddd-book-manager/internal/application/user/dto"
	uq "go-ddd-book-manager/internal/application/user/query"
	dc "go-ddd-book-manager/internal/domain/common"
	"go-ddd-book-manager/internal/webapi/auth"
	"go-ddd-book-manager/internal/webapi/handler"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/samber/do/v2"
)

type Handler struct {
	syncUser    uc.SyncUser
	userDetails uq.GetUserDetails
}

func NewHandler(i do.Injector) (*Handler, error) {
	return &Handler{
		syncUser:    do.MustInvoke[uc.SyncUser](i),
		userDetails: do.MustInvoke[uq.GetUserDetails](i),
	}, nil
}

// sync godoc
// @Summary Synchronize the current user
// @Produce json
// @Success 200 {object} dto.UserResponse
// @Failure 401 {object} handler.ErrorResponse "Unauthorized"
// @Failure 409 {object} handler.ErrorResponse "Conflict"
// @Failure 500 {object} handler.ErrorResponse "Internal Server Error"
// @Header 401 {string} WWW-Authenticate "Bearer"
// @Security BearerAuth
// @Router /users/me [post]
func (h *Handler) sync(c *echo.Context) error {
	return handler.IdentityJSON(
		c,
		http.StatusOK,
		func(ctx context.Context, id auth.Identity) (ud.UserResponse, error) {
			req := ud.SyncUserRequest{ID: id.ID, Name: id.Name, Email: id.Email, Role: id.Role}
			result, err := h.syncUser.Execute(ctx, req)
			if _, ok := errors.AsType[*dc.ValidationError](err); ok {
				return ud.UserResponse{}, auth.ErrUnauthorized
			}
			return result, err
		},
	)
}

// details godoc
// @Summary Get the current user
// @Produce json
// @Success 200 {object} dto.UserResponse
// @Failure 401 {object} handler.ErrorResponse "Unauthorized"
// @Failure 404 {object} handler.ErrorResponse "Not Found"
// @Failure 500 {object} handler.ErrorResponse "Internal Server Error"
// @Header 401 {string} WWW-Authenticate "Bearer"
// @Security BearerAuth
// @Router /users/me [get]
func (h *Handler) details(c *echo.Context) error {
	return handler.JSON(c, http.StatusOK, h.userDetails.Execute)
}
