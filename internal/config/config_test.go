package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"go-ddd-book-manager/internal/config"
)

func TestLoad(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://db")
	t.Setenv("OIDC_ISSUER", "https://identity.example")
	t.Setenv("OIDC_AUDIENCE", "api")
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("CORS_ALLOWED_ORIGINS", " https://app.example.com, ,http://localhost:3000 ,")
	got, err := config.Load()
	require.NoError(t, err)
	require.Equal(t, "postgres://db", got.DatabaseURL)
	require.Equal(t, "https://identity.example", got.OIDCIssuer)
	require.Equal(t, "api", got.OIDCAudience)
	require.Equal(t, ":8080", got.HTTPAddr)
	require.Equal(
		t,
		[]string{"https://app.example.com", "http://localhost:3000"},
		got.AllowedOrigins,
	)
	t.Setenv("HTTP_ADDR", ":9090")
	t.Setenv("CORS_ALLOWED_ORIGINS", " ")
	got, err = config.Load()
	require.NoError(t, err)
	require.Equal(t, ":9090", got.HTTPAddr)
	require.Empty(t, got.AllowedOrigins)
}

func TestLoadRequiresDatabaseAndOIDC(t *testing.T) {
	for _, missing := range []string{"DATABASE_URL", "OIDC_ISSUER", "OIDC_AUDIENCE"} {
		t.Run(missing, func(t *testing.T) {
			t.Setenv("DATABASE_URL", "postgres://db")
			t.Setenv("OIDC_ISSUER", "https://identity.example")
			t.Setenv("OIDC_AUDIENCE", "api")
			t.Setenv(missing, "")
			_, err := config.Load()
			require.Error(t, err)
			require.NoError(t, os.Unsetenv(missing))
			_, err = config.Load()
			require.Error(t, err)
		})
	}
}
