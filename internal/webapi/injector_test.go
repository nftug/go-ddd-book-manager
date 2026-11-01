package webapi_test

import (
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/samber/do/v2"
	"github.com/stretchr/testify/require"
	"go-ddd-book-manager/internal/application"
	"go-ddd-book-manager/internal/config"
	"go-ddd-book-manager/internal/infra"
	"go-ddd-book-manager/internal/webapi"
)

func TestProductionInjection(t *testing.T) {
	i := do.New()
	do.ProvideValue(
		i,
		config.Config{
			DatabaseURL:  "postgres://unused:unused@localhost/unused?sslmode=disable",
			OIDCIssuer:   "https://identity.example/realms/test",
			OIDCAudience: "api",
		},
	)
	infra.Inject(i)
	application.Inject(i)
	webapi.Inject(i)
	t.Cleanup(func() { i.Shutdown() })
	server, err := do.Invoke[*echo.Echo](i)
	require.NoError(t, err)
	require.NotNil(t, server)
}

func TestInvalidOIDCConfigFailsAtStartup(t *testing.T) {
	i := do.New()
	do.ProvideValue(i, config.Config{OIDCIssuer: "invalid", OIDCAudience: "api"})
	webapi.Inject(i)
	t.Cleanup(func() { i.Shutdown() })
	_, err := do.Invoke[*echo.Echo](i)
	require.Error(t, err)
}

func TestInvalidCORSConfigFailsAtStartup(t *testing.T) {
	i := do.New()
	do.ProvideValue(i, config.Config{AllowedOrigins: []string{"not-an-origin"}})
	webapi.Inject(i)
	t.Cleanup(func() { i.Shutdown() })
	_, err := do.Invoke[*echo.Echo](i)
	require.ErrorContains(t, err, "configure CORS")
}
