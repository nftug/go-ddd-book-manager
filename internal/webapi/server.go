package webapi

import (
	us "go-ddd-book-manager/internal/application/user/service"
	"go-ddd-book-manager/internal/webapi/auth"
	"go-ddd-book-manager/internal/webapi/book"
	"go-ddd-book-manager/internal/webapi/handler"
	"go-ddd-book-manager/internal/webapi/user"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/samber/do/v2"
)

func New(i do.Injector) (*echo.Echo, error) {
	verifier := do.MustInvoke[auth.TokenVerifier](i)
	actorResolver := do.MustInvoke[us.ResolveActor](i)
	userHandler := do.MustInvoke[*user.Handler](i)
	bookHandler := do.MustInvoke[*book.Handler](i)

	e := echo.New()
	e.HTTPErrorHandler = handler.ErrorHandler
	e.Use(middleware.RequestID(), middleware.Recover())

	user.RegisterRoutes(e, verifier, actorResolver, userHandler)
	book.RegisterRoutes(e, verifier, actorResolver, bookHandler)

	return e, nil
}
