package infra_test

import (
	"testing"
	"time"

	dc "go-ddd-book-manager/internal/domain/common"
	"go-ddd-book-manager/internal/infra"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/require"
)

func TestInfrastructureServices(t *testing.T) {
	i := infra.NewInjector(do.New(), "postgres://unused:unused@localhost/unused?sslmode=disable")

	clock := do.MustInvoke[dc.Clock](i)
	before := time.Now()
	now := clock.Now()
	require.WithinDuration(t, before, now, time.Second)

	gen := do.MustInvoke[dc.UUIDGenerator](i)
	id := gen.NewUUIDv7()
	require.Equal(t, byte(7), id[6]>>4)
	require.Equal(t, byte(2), id[8]>>6)
	require.NotEqual(t, id, gen.NewUUIDv7())

	i.Shutdown()
}
