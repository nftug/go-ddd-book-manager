package telemetry

import (
	apptelemetry "go-ddd-book-manager/internal/telemetry"
	"time"

	echootel "github.com/labstack/echo-otel/v5"
	"github.com/labstack/echo/v5"
	"go.opentelemetry.io/otel/attribute"
	otellog "go.opentelemetry.io/otel/log"
)

func TracingMiddleware() echo.MiddlewareFunc {
	return echootel.NewMiddlewareWithConfig(echootel.Config{
		SpanStartAttributes: func(
			c *echo.Context, _ *echootel.Values, attrs []attribute.KeyValue) []attribute.KeyValue {
			if query := c.Request().URL.RawQuery; query != "" {
				attrs = append(attrs, attribute.String("url.query", redactQuery(query)))
			}
			return attrs
		},
	})
}

func RequestLogging() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			start := time.Now()
			err := next(c)
			if err != nil {
				c.Echo().HTTPErrorHandler(c, err)
			}
			_, status := echo.ResolveResponseStatus(c.Response(), err)
			apptelemetry.Emit(c.Request().Context(), otellog.SeverityInfo, "request completed",
				attribute.String("request_id", c.Response().Header().Get(echo.HeaderXRequestID)),
				attribute.String("http.request.method", c.Request().Method),
				attribute.String("http.route", c.Path()),
				attribute.String("url.query", redactQuery(c.Request().URL.RawQuery)),
				attribute.Int("http.response.status_code", status),
				attribute.Int64("duration_ms", time.Since(start).Milliseconds()),
			)
			return nil
		}
	}
}
