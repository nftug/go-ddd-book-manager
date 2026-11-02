package telemetry_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"go-ddd-book-manager/internal/telemetry"
	"go.opentelemetry.io/otel"
	otellog "go.opentelemetry.io/otel/log"
)

func TestInitExportsAndFlushes(t *testing.T) {
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
	telemetry.Emit(ctx, otellog.SeverityError, "test message")
	span.End()
	require.NoError(t, shutdown(context.Background()))
	require.Contains(t, stdout.String(), "test message")
	mu.Lock()
	defer mu.Unlock()
	require.NotEmpty(t, paths["/v1/logs"])
	require.NotEmpty(t, paths["/v1/traces"])
}

func TestInitWithoutEndpointWritesStdout(t *testing.T) {
	var stdout bytes.Buffer
	shutdown, err := telemetry.Init(context.Background(), "", &stdout)
	require.NoError(t, err)
	telemetry.Emit(context.Background(), otellog.SeverityInfo, "local only")
	require.NoError(t, shutdown(context.Background()))
	require.Contains(t, stdout.String(), "local only")
}
