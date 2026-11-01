package auth

import (
	us "go-ddd-book-manager/internal/application/user/service"

	"github.com/labstack/echo/v5"
)

func IdentityGroup(e *echo.Echo, prefix string, verifier TokenVerifier) *echo.Group {
	return e.Group(prefix, RequireIdentity(verifier))
}

func ActorGroup(
	e *echo.Echo,
	prefix string,
	verifier TokenVerifier,
	actorResolver us.ResolveActor,
) *echo.Group {
	return e.Group(prefix, RequireIdentity(verifier), RequireUser(actorResolver))
}
