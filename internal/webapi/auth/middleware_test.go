package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"uuid"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/require"
	"go-ddd-book-manager/internal/webapi/auth"
)

type countingVerifier struct {
	calls int
	id    uuid.UUID
	err   error
}

func (v *countingVerifier) Verify(_ context.Context, _ string) (auth.Identity, error) {
	v.calls++
	return auth.Identity{ID: v.id}, v.err
}

func TestRequireIdentity(t *testing.T) {
	id := uuid.MustParse("00000000-0000-4000-8000-000000000001")
	v := &countingVerifier{id: id}
	e := echo.New()
	e.GET("/", func(c *echo.Context) error {
		identity, err := auth.IdentityFrom(c)
		require.NoError(t, err)
		require.Equal(t, id, identity.ID)
		return c.NoContent(http.StatusNoContent)
	}, auth.RequireIdentity(v))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer valid")
	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)
	require.Equal(t, http.StatusNoContent, w.Code)
	require.Equal(t, 1, v.calls)
}
