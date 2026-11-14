package handler_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"uuid"

	"github.com/labstack/echo/v5"
	"github.com/samber/do/v2"
	"github.com/stretchr/testify/require"
	uc "go-ddd-book-manager/internal/application/user/command"
	ud "go-ddd-book-manager/internal/application/user/dto"
	dc "go-ddd-book-manager/internal/domain/common"
	"go-ddd-book-manager/internal/webapi/auth"
	"go-ddd-book-manager/internal/webapi/handler"
)

type verifier struct{ id uuid.UUID }

func (v verifier) Verify(context.Context, string) (auth.Identity, error) {
	return auth.Identity{ID: v.id}, nil
}

type syncFake struct {
	actor dc.Actor
	err   error
}

func (s syncFake) Execute(_ context.Context, req ud.SyncUserRequest) (dc.Actor, error) {
	if s.err != nil {
		return dc.Actor{}, s.err
	}
	return s.actor, nil
}

func newExecutor(t *testing.T, syncer uc.SyncUser) *handler.Executor {
	t.Helper()
	i := do.New()
	do.ProvideValue[uc.SyncUser](i, syncer)
	t.Cleanup(func() { i.Shutdown() })
	executor, err := handler.NewExecutor(i)
	require.NoError(t, err)
	return executor
}

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
			executor := newExecutor(t, syncFake{actor: actor})
			e.HTTPErrorHandler = func(c *echo.Context, err error) {
				require.ErrorIs(t, err, expectedErr)
				_ = c.NoContent(http.StatusTeapot)
			}
			e.GET("/", func(c *echo.Context) error {
				return executor.Created(
					c,
					func(ctx context.Context, got dc.Actor) (any, error) {
						require.Equal(t, actor, got)
						require.Equal(t, "value", ctx.Value(contextKey{}))
						return "result", tc.resultError
					},
				)
			}, auth.RequireIdentity(verifier{id}))
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
	executor := newExecutor(t, syncFake{})
	c := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), httptest.NewRecorder())
	called := false
	err := executor.NoContent(
		c,
		func(context.Context, dc.Actor) error { called = true; return nil },
	)
	require.Error(t, err)
	require.False(t, called)
	errResult := executor.OK(
		c,
		func(context.Context, dc.Actor) (any, error) { called = true; return "", nil },
	)
	require.Error(t, errResult)
	require.False(t, called)
	id := uuid.MustParse("00000000-0000-4000-8000-000000000001")
	actor := dc.NewActor(id, "User", dc.UserRole)
	expected := errors.New("failed")
	executor = newExecutor(t, syncFake{actor: actor})
	e.POST("/", func(c *echo.Context) error {
		return executor.NoContent(
			c,
			func(ctx context.Context, got dc.Actor) error {
				require.Equal(t, actor, got)
				require.Equal(t, "value", ctx.Value(contextKey{}))
				return nil
			},
		)
	}, auth.RequireIdentity(verifier{id}))
	e.POST("/error", func(c *echo.Context) error {
		return executor.NoContent(
			c,
			func(context.Context, dc.Actor) error { return expected },
		)
	}, auth.RequireIdentity(verifier{id}))
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

func TestOKReturnsJSON(t *testing.T) {
	id := uuid.MustParse("00000000-0000-4000-8000-000000000001")
	executor := newExecutor(t, syncFake{actor: dc.NewActor(id, "User", dc.UserRole)})
	e := echo.New()
	e.GET("/", func(c *echo.Context) error {
		return executor.OK(c, func(_ context.Context, actor dc.Actor) (any, error) {
			return map[string]string{"id": actor.ID().String()}, nil
		})
	}, auth.RequireIdentity(verifier{id}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer valid")
	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	require.JSONEq(t, `{"id":"`+id.String()+`"}`, w.Body.String())
}
