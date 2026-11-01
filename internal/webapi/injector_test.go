package webapi_test

import (
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/samber/do/v2"
	"github.com/stretchr/testify/require"
	"go-ddd-book-manager/internal/application"
	"go-ddd-book-manager/internal/infra"
	"go-ddd-book-manager/internal/webapi"
	"go-ddd-book-manager/internal/webapi/auth"
)

func TestProductionInjection(t *testing.T) {
	i := webapi.Inject(
		application.Inject(
			infra.Inject(do.New(), "postgres://unused:unused@localhost/unused?sslmode=disable"),
		),
		auth.Config{Issuer: "https://identity.example/realms/test", Audience: "api"},
		webapi.Config{},
	)
	t.Cleanup(func() { i.Shutdown() })
	server, err := do.Invoke[*echo.Echo](i)
	require.NoError(t, err)
	require.NotNil(t, server)
}

func TestInvalidOIDCConfigFailsAtStartup(t *testing.T) {
	i := webapi.Inject(do.New(), auth.Config{Issuer: "invalid", Audience: "api"}, webapi.Config{})
	t.Cleanup(func() { i.Shutdown() })
	_, err := do.Invoke[*echo.Echo](i)
	require.Error(t, err)
}

func TestInvalidCORSConfigFailsAtStartup(t *testing.T) {
	i := webapi.Inject(
		do.New(),
		auth.Config{},
		webapi.Config{AllowedOrigins: []string{"not-an-origin"}},
	)
	t.Cleanup(func() { i.Shutdown() })
	_, err := do.Invoke[*echo.Echo](i)
	require.ErrorContains(t, err, "configure CORS")
}
