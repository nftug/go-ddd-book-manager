package auth_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"uuid"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/require"
	dc "go-ddd-book-manager/internal/domain/common"
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

type countingResolver struct {
	calls int
	id    uuid.UUID
	err   error
}

func (r *countingResolver) Execute(_ context.Context, id uuid.UUID) (dc.Actor, error) {
	r.calls++
	r.id = id
	return dc.NewActor(id, "User", dc.UserRole), r.err
}

func TestRequireUserIncludesIdentity(t *testing.T) {
	id := uuid.MustParse("00000000-0000-4000-8000-000000000001")
	providerErr := errors.New("provider error")
	for _, tc := range []struct {
		name, header                          string
		verifyErr, resolveErr                 error
		wantStatus, verifyCalls, resolveCalls int
	}{
		{"valid", "Bearer valid", nil, nil, http.StatusNoContent, 1, 1},
		{"missing token", "", nil, nil, http.StatusUnauthorized, 0, 0},
		{"invalid token", "Bearer invalid", auth.ErrUnauthorized, nil, http.StatusUnauthorized, 1, 0},
		{"unknown user", "Bearer valid", nil, dc.ErrNotFound, http.StatusUnauthorized, 1, 1},
		{"provider failure", "Bearer valid", providerErr, nil, http.StatusInternalServerError, 1, 0},
		{"resolver failure", "Bearer valid", nil, providerErr, http.StatusInternalServerError, 1, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			v := &countingVerifier{id: id, err: tc.verifyErr}
			r := &countingResolver{err: tc.resolveErr}
			e := echo.New()
			e.HTTPErrorHandler = func(c *echo.Context, err error) {
				status := http.StatusInternalServerError
				if errors.Is(err, auth.ErrUnauthorized) {
					status = http.StatusUnauthorized
				}
				_ = c.NoContent(status)
			}
			e.GET("/", func(c *echo.Context) error {
				identity, err := auth.IdentityFrom(c)
				require.NoError(t, err)
				require.Equal(t, id, identity.ID)
				actor, err := auth.ActorFrom(c)
				require.NoError(t, err)
				require.Equal(t, id, actor.ID())
				return c.NoContent(http.StatusNoContent)
			}, auth.RequireUser(v, r))
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.header != "" {
				req.Header.Set(echo.HeaderAuthorization, tc.header)
			}
			w := httptest.NewRecorder()
			e.ServeHTTP(w, req)
			require.Equal(t, tc.wantStatus, w.Code)
			require.Equal(t, tc.verifyCalls, v.calls)
			require.Equal(t, tc.resolveCalls, r.calls)
			if r.calls > 0 {
				require.Equal(t, id, r.id)
			}
		})
	}
}
