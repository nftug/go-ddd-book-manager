//go:build integration

package infra_test

import (
	"context"
	"github.com/stretchr/testify/require"
	"go-ddd-book-manager/internal/infra/testutil"
	"testing"
)

func TestPostgresMigrationsAndShutdown(t *testing.T) {
	ctx := context.Background()
	i, db, m := testutil.NewPostgres(t)

	// ロールバックで作成したテーブルが削除される
	_, err := m.Rollback(ctx)
	require.NoError(t, err)
	exists, err := db.NewSelect().Table("information_schema.tables").
		Where("table_schema = current_schema() AND table_name = ?", "books").Exists(ctx)
	require.NoError(t, err)
	require.False(t, exists)

	// Injector の終了でデータベース接続も閉じる
	i.Shutdown()
	require.Error(t, db.PingContext(ctx))
}
