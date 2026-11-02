package telemetry_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"uuid"

	"github.com/samber/do/v2"
	us "go-ddd-book-manager/internal/application/user/service"
	"go-ddd-book-manager/internal/config"
	dc "go-ddd-book-manager/internal/domain/common"
	"go-ddd-book-manager/internal/webapi"
	"go-ddd-book-manager/internal/webapi/auth"
	"go-ddd-book-manager/internal/webapi/book"
	"go-ddd-book-manager/internal/webapi/user"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/require"
	"go-ddd-book-manager/internal/telemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestRequestTelemetry(t *testing.T) {
	old := slog.Default()
	t.Cleanup(func() { slog.SetDefault(old) })
	var stdout bytes.Buffer
	shutdown, err := telemetry.Init(context.Background(), "", &stdout)
	require.NoError(t, err)
	defer func() { require.NoError(t, shutdown(context.Background())) }()
	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	otel.SetTracerProvider(provider)
	defer func() { require.NoError(t, provider.Shutdown(context.Background())) }()
	i := do.New()
	defer i.Shutdown()
	do.ProvideValue(i, config.Config{})
	do.ProvideValue[auth.TokenVerifier](i, rejectingVerifier{})
	do.ProvideValue[us.ResolveActor](i, unusedResolver{})
	do.ProvideValue(i, &user.Handler{})
	do.ProvideValue(i, &book.Handler{})
	e, err := webapi.New(i)
	require.NoError(t, err)
	e.GET("/telemetry/ok", func(c *echo.Context) error {
		require.Equal(
			t,
			"testvalue",
			baggage.FromContext(c.Request().Context()).Member("testkey").Value(),
		)
		return c.NoContent(http.StatusNoContent)
	})
	e.GET("/telemetry/domain", func(c *echo.Context) error { return dc.ErrForbidden })
	e.GET("/telemetry/committed", func(c *echo.Context) error {
		_ = c.String(http.StatusAccepted, "sent")
		return errors.New("after response")
	})
	e.GET("/telemetry/committed-client-error", func(c *echo.Context) error {
		_ = c.String(http.StatusBadRequest, "sent")
		return errors.New("after response")
	})
	e.GET("/telemetry/fail", func(c *echo.Context) error { return errors.New("failure") })
	e.GET("/telemetry/panic", func(c *echo.Context) error { panic("failure") })
	for _, tc := range []struct {
		path   string
		status int
	}{{"/telemetry/ok", 204}, {"/users/me", 401}, {"/missing", 404}, {"/telemetry/domain", 403}, {"/telemetry/committed", 202}, {"/telemetry/committed-client-error", 400}, {"/telemetry/fail", 500}, {"/telemetry/panic", 500}} {
		t.Run(tc.path, func(t *testing.T) {
			stdout.Reset()
			req := httptest.NewRequest(
				http.MethodGet,
				tc.path+"?q=hello&token=topsecret&password=secretpass",
				nil,
			)
			req.Header.Set("Authorization", "Bearer secret-auth")
			req.Header.Set("baggage", "testkey=testvalue")
			req.Header.Set("traceparent", "00-0123456789abcdef0123456789abcdef-0123456789abcdef-01")
			w := httptest.NewRecorder()
			e.ServeHTTP(w, req)
			require.Equal(t, tc.status, w.Code)
			out := stdout.String()
			require.Contains(t, out, "request completed")
			require.Equal(t, 1, strings.Count(out, "request completed"))
			require.NotContains(t, out, "request failed")
			var logged map[string]any
			require.NoError(t, json.Unmarshal(stdout.Bytes(), &logged))
			require.Equal(t, "request completed", logged["msg"])
			switch tc.status {
			case 204:
				require.Equal(t, "INFO", logged["level"])
				require.NotContains(t, logged, "error")
				require.Empty(t, w.Body.String())
			case 401, 403, 404:
				require.Equal(t, "WARN", logged["level"])
				require.NotContains(t, logged, "error")
				require.JSONEq(
					t,
					fmt.Sprintf(`{"message":%q}`, http.StatusText(tc.status)),
					w.Body.String(),
				)
				if tc.status == 401 {
					require.Equal(t, "Bearer", w.Header().Get(echo.HeaderWWWAuthenticate))
				}
			case 202:
				require.Equal(t, "ERROR", logged["level"])
				require.Equal(t, "after response", logged["error"])
				require.Equal(t, "sent", w.Body.String())
			case 400:
				require.Equal(t, "WARN", logged["level"])
				require.NotContains(t, logged, "error")
				require.Equal(t, "sent", w.Body.String())
			case 500:
				require.Equal(t, "ERROR", logged["level"])
				require.Contains(t, logged["error"], "failure")
				require.JSONEq(t, `{"message":"Internal Server Error"}`, w.Body.String())
			}
			require.NotEmpty(t, logged["request_id"])
			require.Equal(t, "GET", logged["http.request.method"])
			require.Contains(t, logged, "http.route")
			require.Contains(t, logged, "duration_ms")
			require.Contains(t, out, fmt.Sprintf(`"http.response.status_code":%d`, tc.status))
			require.Contains(t, out, "q=hello")
			require.NotContains(t, out, "topsecret")
			require.NotContains(t, out, "secretpass")
			require.NotContains(t, out, "secret-auth")
			spans := recorder.Ended()
			span := spans[len(spans)-1]
			require.True(t, span.SpanContext().IsValid())
			if tc.status == 500 {
				require.Equal(t, codes.Error, span.Status().Code)
			}
			require.Equal(
				t,
				"0123456789abcdef0123456789abcdef",
				span.SpanContext().TraceID().String(),
			)
			require.Equal(t, "0123456789abcdef", span.Parent().SpanID().String())
			attrs := span.Attributes()
			require.Contains(t, attrs, attribute.Int("http.response.status_code", tc.status))
			require.Contains(
				t,
				attrs,
				attribute.String("url.query", "q=hello&token=REDACTED&password=REDACTED"),
			)
		})
	}
}

// These dependencies are only needed to construct the real router; the protected
// route rejects the request before reaching its handler or resolving an actor.
type rejectingVerifier struct{}

func (rejectingVerifier) Verify(context.Context, string) (auth.Identity, error) {
	return auth.Identity{}, auth.ErrUnauthorized
}

type unusedResolver struct{}

func (unusedResolver) Execute(context.Context, uuid.UUID) (dc.Actor, error) {
	panic("unexpected actor resolution")
}
