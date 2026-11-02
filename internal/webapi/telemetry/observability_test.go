package telemetry_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
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
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestRequestTelemetry(t *testing.T) {
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
	e.GET("/telemetry/ok", func(c *echo.Context) error { return c.NoContent(http.StatusNoContent) })
	e.GET("/telemetry/fail", func(c *echo.Context) error { return errors.New("failure") })
	e.GET("/telemetry/panic", func(c *echo.Context) error { panic("failure") })
	for _, tc := range []struct {
		path   string
		status int
	}{{"/telemetry/ok", 204}, {"/users/me", 401}, {"/missing", 404}, {"/telemetry/fail", 500}, {"/telemetry/panic", 500}} {
		t.Run(tc.path, func(t *testing.T) {
			stdout.Reset()
			req := httptest.NewRequest(
				http.MethodGet,
				tc.path+"?q=hello&token=topsecret&password=secretpass",
				nil,
			)
			req.Header.Set("Authorization", "Bearer secret-auth")
			req.Header.Set("traceparent", "00-0123456789abcdef0123456789abcdef-0123456789abcdef-01")
			w := httptest.NewRecorder()
			e.ServeHTTP(w, req)
			require.Equal(t, tc.status, w.Code)
			out := stdout.String()
			require.Contains(t, out, "request completed")
			require.Contains(t, out, "q=hello")
			require.NotContains(t, out, "topsecret")
			require.NotContains(t, out, "secretpass")
			require.NotContains(t, out, "secret-auth")
			spans := recorder.Ended()
			span := spans[len(spans)-1]
			require.True(t, span.SpanContext().IsValid())
			require.Equal(
				t,
				"0123456789abcdef0123456789abcdef",
				span.SpanContext().TraceID().String(),
			)
			require.Equal(t, "0123456789abcdef", span.Parent().SpanID().String())
			require.Contains(t, out, span.SpanContext().TraceID().String())
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
