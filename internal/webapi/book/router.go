package book

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
	routes := e.Group("/books", auth.RequireUser(verifier, actorResolver))
	routes.GET("", h.list)
	routes.POST("", h.create)
	routes.GET("/:book_id", h.details)
	routes.PUT("/:book_id", h.update)
	routes.DELETE("/:book_id", h.delete)
	routes.PUT("/:book_id/owner", h.changeOwner)
	routes.POST("/:book_id/checkouts", h.checkout)
	routes.POST("/:book_id/return", h.returnBook)
	routes.GET("/:book_id/checkouts", h.history)
}
