package telemetry

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"strings"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Init installs the providers and returns a function that flushes and shuts them down.
func Init(
	ctx context.Context,
	endpoint string,
	output io.Writer,
) (func(context.Context) error, error) {
	lp, tp, err := newProviders(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	stdout := slog.NewJSONHandler(output, &slog.HandlerOptions{ReplaceAttr: utcTimestamp})
	handlers := []slog.Handler{stdout}
	if endpoint != "" {
		handlers = append(
			handlers,
			otelslog.NewHandler("go-ddd-book-manager", otelslog.WithLoggerProvider(lp)),
		)
	}
	slog.SetDefault(slog.New(slog.NewMultiHandler(handlers...)))
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)
	return func(ctx context.Context) error { return errors.Join(tp.Shutdown(ctx), lp.Shutdown(ctx)) }, nil
}

func newProviders(
	ctx context.Context,
	endpoint string,
) (*sdklog.LoggerProvider, *sdktrace.TracerProvider, error) {
	res := resource.NewWithAttributes("", attribute.String("service.name", "go-ddd-book-manager"))
	if endpoint == "" {
		lp := sdklog.NewLoggerProvider(sdklog.WithResource(res))
		tp := sdktrace.NewTracerProvider(sdktrace.WithResource(res))
		return lp, tp, nil
	}

	base, err := url.Parse(endpoint)
	if err != nil || (base.Scheme != "http" && base.Scheme != "https") || base.Host == "" ||
		base.RawQuery != "" ||
		base.Fragment != "" {
		return nil, nil, fmt.Errorf("invalid OTLP endpoint %q", endpoint)
	}
	base.Path = strings.TrimRight(base.Path, "/")
	logsURL := *base
	logsURL.Path += "/v1/logs"
	tracesURL := *base
	tracesURL.Path += "/v1/traces"
	logExporter, err := otlploghttp.New(ctx, otlploghttp.WithEndpointURL(logsURL.String()))
	if err != nil {
		return nil, nil, fmt.Errorf("configure OTLP log exporter: %w", err)
	}
	traceExporter, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpointURL(tracesURL.String()))
	if err != nil {
		_ = logExporter.Shutdown(ctx)
		return nil, nil, fmt.Errorf("configure OTLP trace exporter: %w", err)
	}

	lp := sdklog.NewLoggerProvider(
		sdklog.WithResource(res),
		sdklog.WithProcessor(sdklog.NewBatchProcessor(logExporter)),
	)
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(traceExporter),
	)
	return lp, tp, nil
}

func utcTimestamp(_ []string, a slog.Attr) slog.Attr {
	if a.Key == slog.TimeKey && a.Value.Kind() == slog.KindTime {
		return slog.Time(a.Key, a.Value.Time().UTC())
	}
	return a
}
