package telemetry

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	otellog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
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
	stdout, err := stdoutlog.New(stdoutlog.WithWriter(output))
	if err != nil {
		return nil, fmt.Errorf("configure stdout log exporter: %w", err)
	}
	res := resource.NewWithAttributes("", attribute.String("service.name", "go-ddd-book-manager"))
	logOptions := []sdklog.LoggerProviderOption{
		sdklog.WithResource(res),
		sdklog.WithProcessor(sdklog.NewSimpleProcessor(stdout)),
	}
	traceOptions := []sdktrace.TracerProviderOption{sdktrace.WithResource(res)}
	if endpoint != "" {
		base, err := url.Parse(endpoint)
		if err != nil || (base.Scheme != "http" && base.Scheme != "https") || base.Host == "" ||
			base.RawQuery != "" ||
			base.Fragment != "" {
			return nil, fmt.Errorf("invalid OTLP endpoint %q", endpoint)
		}
		base.Path = strings.TrimRight(base.Path, "/")
		logsURL := *base
		logsURL.Path += "/v1/logs"
		tracesURL := *base
		tracesURL.Path += "/v1/traces"
		logExporter, err := otlploghttp.New(ctx, otlploghttp.WithEndpointURL(logsURL.String()))
		if err != nil {
			return nil, fmt.Errorf("configure OTLP log exporter: %w", err)
		}
		traceExporter, err := otlptracehttp.New(
			ctx,
			otlptracehttp.WithEndpointURL(tracesURL.String()),
		)
		if err != nil {
			_ = logExporter.Shutdown(ctx)
			return nil, fmt.Errorf("configure OTLP trace exporter: %w", err)
		}
		logOptions = append(logOptions, sdklog.WithProcessor(sdklog.NewBatchProcessor(logExporter)))
		traceOptions = append(traceOptions, sdktrace.WithBatcher(traceExporter))
	}
	lp := sdklog.NewLoggerProvider(logOptions...)
	tp := sdktrace.NewTracerProvider(traceOptions...)
	global.SetLoggerProvider(lp)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)
	return func(ctx context.Context) error { return errors.Join(tp.Shutdown(ctx), lp.Shutdown(ctx)) }, nil
}

func Emit(
	ctx context.Context,
	severity otellog.Severity,
	message string,
	attrs ...attribute.KeyValue,
) {
	var record otellog.Record
	record.SetTimestamp(time.Now().UTC())
	record.SetSeverity(severity)
	record.SetSeverityText(severity.String())
	record.SetBody(attribute.StringValue(message))
	record.AddAttributes(attrs...)
	global.Logger("go-ddd-book-manager").Emit(ctx, record)
}
