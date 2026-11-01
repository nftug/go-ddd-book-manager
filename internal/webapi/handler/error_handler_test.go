package handler

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	dc "go-ddd-book-manager/internal/domain/common"
	"go-ddd-book-manager/internal/webapi/auth"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/require"
)

func TestErrorHandlerResponses(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
		body   string
	}{
		{
			"unauthorized",
			fmt.Errorf("wrapped: %w", auth.ErrUnauthorized),
			401,
			`{"message":"Unauthorized"}`,
		},
		{"forbidden", fmt.Errorf("wrapped: %w", dc.ErrForbidden), 403, `{"message":"Forbidden"}`},
		{"not found", fmt.Errorf("wrapped: %w", dc.ErrNotFound), 404, `{"message":"Not Found"}`},
		{"conflict", fmt.Errorf("wrapped: %w", dc.ErrConflict), 409, `{"message":"Conflict"}`},
		{
			"validation",
			fmt.Errorf("wrapped: %w", dc.NewValidationError("title", "invalid")),
			400,
			`{"message":"Invalid input","field":"title","code":"invalid"}`,
		},
		{
			"validation with empty details",
			dc.NewValidationError("", ""),
			400,
			`{"message":"Invalid input","field":"","code":""}`,
		},
		{
			"echo error",
			fmt.Errorf("wrapped: %w", echo.NewHTTPError(415, "private detail")),
			415,
			`{"message":"Unsupported Media Type"}`,
		},
		{
			"echo server error",
			echo.NewHTTPError(503, "private detail"),
			503,
			`{"message":"Internal Server Error"}`,
		},
		{
			"unexpected error",
			errors.New("private detail"),
			500,
			`{"message":"Internal Server Error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			e.HTTPErrorHandler = ErrorHandler
			e.GET("/", func(*echo.Context) error { return tt.err })
			w := httptest.NewRecorder()
			e.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
			require.Equal(t, tt.status, w.Code)
			require.JSONEq(t, tt.body, w.Body.String())
			if tt.status == http.StatusUnauthorized {
				require.Equal(t, "Bearer", w.Header().Get(echo.HeaderWWWAuthenticate))
			} else {
				require.Empty(t, w.Header().Get(echo.HeaderWWWAuthenticate))
			}
		})
	}
}

func TestErrorHandlerHeadAndCommittedResponse(t *testing.T) {
	e := echo.New()
	e.HTTPErrorHandler = ErrorHandler
	e.HEAD("/", func(*echo.Context) error { return auth.ErrUnauthorized })
	e.GET("/", func(c *echo.Context) error {
		if err := c.String(http.StatusAccepted, "already sent"); err != nil {
			return err
		}
		return errors.New("private detail")
	})

	w := httptest.NewRecorder()
	e.ServeHTTP(w, httptest.NewRequest(http.MethodHead, "/", nil))
	require.Equal(t, http.StatusUnauthorized, w.Code)
	require.Equal(t, "Bearer", w.Header().Get(echo.HeaderWWWAuthenticate))
	require.Empty(t, w.Body.String())

	w = httptest.NewRecorder()
	e.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
	require.Equal(t, http.StatusAccepted, w.Code)
	require.Equal(t, "already sent", w.Body.String())
}

func TestErrorHandlerRouterErrors(t *testing.T) {
	e := echo.New()
	e.HTTPErrorHandler = ErrorHandler
	e.GET("/registered", func(c *echo.Context) error { return c.NoContent(http.StatusNoContent) })
	for _, tt := range []struct {
		method, path string
		status       int
	}{
		{http.MethodGet, "/missing", http.StatusNotFound},
		{http.MethodPost, "/registered", http.StatusMethodNotAllowed},
	} {
		t.Run(tt.method+tt.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			e.ServeHTTP(w, httptest.NewRequest(tt.method, tt.path, nil))
			require.Equal(t, tt.status, w.Code)
			require.JSONEq(
				t,
				fmt.Sprintf(`{"message":%q}`, http.StatusText(tt.status)),
				w.Body.String(),
			)
		})
	}
}
