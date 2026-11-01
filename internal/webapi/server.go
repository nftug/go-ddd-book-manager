package webapi

import (
	"fmt"
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
	config := do.MustInvoke[Config](i)
	var cors echo.MiddlewareFunc
	if len(config.AllowedOrigins) > 0 {
		var err error
		cors, err = (middleware.CORSConfig{
			AllowOrigins: config.AllowedOrigins,
			AllowHeaders: []string{echo.HeaderAuthorization, echo.HeaderContentType},
		}).ToMiddleware()
		if err != nil {
			return nil, fmt.Errorf("configure CORS: %w", err)
		}
	}
	verifier := do.MustInvoke[auth.TokenVerifier](i)
	actorResolver := do.MustInvoke[us.ResolveActor](i)
	userHandler := do.MustInvoke[*user.Handler](i)
	bookHandler := do.MustInvoke[*book.Handler](i)

	e := echo.New()
	e.HTTPErrorHandler = handler.ErrorHandler
	e.Use(middleware.RequestID(), middleware.Recover())
	if cors != nil {
		e.Use(cors)
	}

	user.RegisterRoutes(e, verifier, actorResolver, userHandler)
	book.RegisterRoutes(e, verifier, actorResolver, bookHandler)

	return e, nil
}
