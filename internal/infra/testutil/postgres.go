//go:build integration

package testutil

import (
	"context"
	"fmt"
	"github.com/ory/dockertest/v3"
	"github.com/samber/do/v2"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/migrate"
	"go-ddd-book-manager/internal/infra"
	"go-ddd-book-manager/internal/infra/migrations"
	"testing"
)

func NewPostgres(t *testing.T) (do.Injector, *bun.DB, *migrate.Migrator) {
	t.Helper()

	// PostgreSQL コンテナを起動し、接続できるまで待つ
	ctx := context.Background()
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)
	require.NoError(t, pool.Client.Ping())
	resource, err := pool.Run(
		"postgres",
		"18-alpine",
		[]string{"POSTGRES_USER=test", "POSTGRES_PASSWORD=test", "POSTGRES_DB=test"},
	)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, pool.Purge(resource)) })
	dsn := fmt.Sprintf(
		"postgres://test:test@localhost:%s/test?sslmode=disable",
		resource.GetPort("5432/tcp"),
	)
	i := infra.NewInjector(do.New(), dsn)
	db := do.MustInvoke[*bun.DB](i)
	t.Cleanup(func() { i.Shutdown() })
	require.NoError(t, pool.Retry(func() error { return db.PingContext(ctx) }))

	// テスト用データベースにすべてのマイグレーションを適用する
	m, err := migrations.New(db)
	require.NoError(t, err)
	require.NoError(t, m.Init(ctx))
	_, err = m.Migrate(ctx)
	require.NoError(t, err)
	return i, db, m
}
