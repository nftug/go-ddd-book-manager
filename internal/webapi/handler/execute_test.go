package handler_test

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
	"go-ddd-book-manager/internal/webapi/handler"
)

type verifier struct{ id uuid.UUID }

func (v verifier) Verify(context.Context, string) (auth.Identity, error) {
	return auth.Identity{ID: v.id}, nil
}

type resolver struct{ actor dc.Actor }

func (r resolver) Execute(context.Context, uuid.UUID) (dc.Actor, error) { return r.actor, nil }

type contextKey struct{}

func TestExecuteWithActor(t *testing.T) {
	id := uuid.MustParse("00000000-0000-4000-8000-000000000001")
	actor := dc.NewActor(id, "User", dc.UserRole)
	expectedErr := errors.New("use case error")
	for _, tc := range []struct {
		name        string
		resultError error
		status      int
	}{{"success", nil, http.StatusCreated}, {"error", expectedErr, http.StatusTeapot}} {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			e.HTTPErrorHandler = func(c *echo.Context, err error) {
				require.ErrorIs(t, err, expectedErr)
				_ = c.NoContent(http.StatusTeapot)
			}
			e.GET("/", func(c *echo.Context) error {
				return handler.JSON(
					c,
					http.StatusCreated,
					func(ctx context.Context, got dc.Actor) (string, error) {
						require.Equal(t, actor, got)
						require.Equal(t, "value", ctx.Value(contextKey{}))
						return "result", tc.resultError
					},
				)
			}, auth.RequireIdentity(verifier{id}), auth.RequireUser(resolver{actor}))
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Authorization", "Bearer valid")
			req = req.WithContext(context.WithValue(req.Context(), contextKey{}, "value"))
			w := httptest.NewRecorder()
			e.ServeHTTP(w, req)
			require.Equal(t, tc.status, w.Code)
			if tc.resultError == nil {
				require.JSONEq(t, `"result"`, w.Body.String())
			}
		})
	}
}

func TestNoContentAndMissingActor(t *testing.T) {
	e := echo.New()
	c := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), httptest.NewRecorder())
	called := false
	err := handler.NoContent(c, func(context.Context, dc.Actor) error { called = true; return nil })
	require.Error(t, err)
	require.False(t, called)
	errResult := handler.JSON(
		c,
		http.StatusOK,
		func(context.Context, dc.Actor) (string, error) { called = true; return "", nil },
	)
	require.Error(t, errResult)
	require.False(t, called)
	id := uuid.MustParse("00000000-0000-4000-8000-000000000001")
	actor := dc.NewActor(id, "User", dc.UserRole)
	expected := errors.New("failed")
	e.POST("/", func(c *echo.Context) error {
		return handler.NoContent(c, func(ctx context.Context, got dc.Actor) error {
			require.Equal(t, actor, got)
			require.Equal(t, "value", ctx.Value(contextKey{}))
			return nil
		})
	}, auth.RequireIdentity(verifier{id}), auth.RequireUser(resolver{actor}))
	e.POST("/error", func(c *echo.Context) error {
		return handler.NoContent(c, func(context.Context, dc.Actor) error { return expected })
	}, auth.RequireIdentity(verifier{id}), auth.RequireUser(resolver{actor}))
	for _, tc := range []struct {
		path   string
		status int
	}{{"/", 204}, {"/error", 500}} {
		req := httptest.NewRequest(http.MethodPost, tc.path, nil)
		req.Header.Set("Authorization", "Bearer valid")
		req = req.WithContext(context.WithValue(req.Context(), contextKey{}, "value"))
		w := httptest.NewRecorder()
		e.ServeHTTP(w, req)
		require.Equal(t, tc.status, w.Code)
		if tc.status == 204 {
			require.Empty(t, w.Body.String())
		}
	}
}

func TestIdentityJSON(t *testing.T) {
	id := uuid.MustParse("00000000-0000-4000-8000-000000000001")
	expectedErr := errors.New("sync failed")
	for _, tc := range []struct {
		name   string
		err    error
		status int
	}{{"success", nil, http.StatusOK}, {"error", expectedErr, http.StatusTeapot}} {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			e.HTTPErrorHandler = func(c *echo.Context, err error) {
				require.ErrorIs(t, err, expectedErr)
				_ = c.NoContent(http.StatusTeapot)
			}
			e.POST("/", func(c *echo.Context) error {
				return handler.IdentityJSON(
					c,
					http.StatusOK,
					func(ctx context.Context, identity auth.Identity) (string, error) {
						require.Equal(t, id, identity.ID)
						require.Equal(t, "value", ctx.Value(contextKey{}))
						return "synced", tc.err
					},
				)
			}, auth.RequireIdentity(verifier{id}))
			req := httptest.NewRequest(http.MethodPost, "/", nil)
			req.Header.Set("Authorization", "Bearer valid")
			req = req.WithContext(context.WithValue(req.Context(), contextKey{}, "value"))
			w := httptest.NewRecorder()
			e.ServeHTTP(w, req)
			require.Equal(t, tc.status, w.Code)
			if tc.err == nil {
				require.JSONEq(t, `"synced"`, w.Body.String())
			}
		})
	}
	c := echo.New().
		NewContext(httptest.NewRequest(http.MethodPost, "/", nil), httptest.NewRecorder())
	called := false
	err := handler.IdentityJSON(
		c,
		http.StatusOK,
		func(context.Context, auth.Identity) (string, error) { called = true; return "", nil },
	)
	require.Error(t, err)
	require.False(t, called)
}
