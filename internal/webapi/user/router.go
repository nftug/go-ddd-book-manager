package user

import (
	us "go-ddd-book-manager/internal/application/user/service"
	"go-ddd-book-manager/internal/webapi/auth"

	"github.com/labstack/echo/v5"
)

func RegisterRoutes(
	e *echo.Echo,
	verifier auth.TokenVerifier,
	actorResolver us.ResolveActor,
	h *Handler,
) {
	idRoutes := e.Group("/users", auth.RequireIdentity(verifier))
	actorRoutes := e.Group("/users", auth.RequireUser(verifier, actorResolver))
	idRoutes.POST("/me", h.sync)
	actorRoutes.GET("/me", h.details)
}
