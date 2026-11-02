package telemetry

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRedactQuery(t *testing.T) {
	require.Equal(
		t,
		"q=golang&token=REDACTED&%70assword=REDACTED&client_secret=REDACTED&access_token=REDACTED&TOKEN=REDACTED",
		redactQuery(
			"q=golang&token=one&%70assword=two&client_secret=three&access_token=four&TOKEN=five",
		),
	)
	require.Equal(t, "[REDACTED]", redactQuery("%zz=secret"))
}
