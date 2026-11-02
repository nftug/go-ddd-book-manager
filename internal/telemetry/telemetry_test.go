package telemetry_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go-ddd-book-manager/internal/telemetry"
	"go.opentelemetry.io/otel"
	collectorlog "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	"google.golang.org/protobuf/proto"
)

func TestInitExportsAndFlushes(t *testing.T) {
	old := slog.Default()
	t.Cleanup(func() { slog.SetDefault(old) })
	var mu sync.Mutex
	paths := map[string][]byte{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		mu.Lock()
		paths[r.URL.Path] = body
		mu.Unlock()
		w.Header().Set("Content-Type", "application/x-protobuf")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	var stdout bytes.Buffer
	shutdown, err := telemetry.Init(context.Background(), server.URL, &stdout)
	require.NoError(t, err)
	ctx, span := otel.Tracer("test").Start(context.Background(), "test span")
	slog.ErrorContext(ctx, "test message", "request_id", "request-123")
	span.End()
	require.NoError(t, shutdown(context.Background()))
	require.Contains(t, stdout.String(), "test message")
	require.Contains(t, stdout.String(), `"request_id":"request-123"`)
	mu.Lock()
	defer mu.Unlock()
	require.NotEmpty(t, paths["/v1/logs"])
	var payload collectorlog.ExportLogsServiceRequest
	require.NoError(t, proto.Unmarshal(paths["/v1/logs"], &payload))
	record := payload.ResourceLogs[0].ScopeLogs[0].LogRecords[0]
	require.Equal(t, "test message", record.Body.GetStringValue())
	require.Equal(t, span.SpanContext().TraceID().String(), fmt.Sprintf("%x", record.TraceId))
	require.NotEmpty(t, paths["/v1/traces"])
}

func TestInitWithoutEndpointWritesStdout(t *testing.T) {
	old := slog.Default()
	t.Cleanup(func() { slog.SetDefault(old) })
	var stdout bytes.Buffer
	shutdown, err := telemetry.Init(context.Background(), "", &stdout)
	require.NoError(t, err)
	slog.InfoContext(context.Background(), "local only", "request_id", "request-456")
	require.NoError(t, shutdown(context.Background()))
	require.Contains(t, stdout.String(), "local only")
	require.Contains(t, stdout.String(), `"request_id":"request-456"`)
	var logged struct {
		Time time.Time `json:"time"`
	}
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &logged))
	_, offset := logged.Time.Zone()
	require.Zero(t, offset)
}
