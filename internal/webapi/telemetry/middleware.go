package telemetry

import (
	"log/slog"

	echootel "github.com/labstack/echo-otel/v5"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"go.opentelemetry.io/otel/attribute"
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
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		HandleError:  true,
		LogRequestID: true,
		LogMethod:    true,
		LogRoutePath: true,
		LogStatus:    true,
		LogLatency:   true,
		LogValuesFunc: func(c *echo.Context, v middleware.RequestLoggerValues) error {
			attrs := []any{
				"request_id", v.RequestID,
				"http.request.method", v.Method,
				"http.route", v.RoutePath,
				"url.query", redactQuery(c.Request().URL.RawQuery),
				"http.response.status_code", v.Status,
				"duration_ms", v.Latency.Milliseconds(),
			}
			level := slog.LevelInfo
			if v.Error != nil {
				if v.Status >= 500 || v.Status < 400 {
					level = slog.LevelError
					attrs = append(attrs, "error", v.Error)
				} else {
					level = slog.LevelWarn
				}
			}
			slog.Log(c.Request().Context(), level, "request completed", attrs...)
			return nil
		},
	})
}
