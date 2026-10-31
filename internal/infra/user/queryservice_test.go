//go:build integration

package user_test

import (
	"context"
	"github.com/samber/do/v2"
	"github.com/stretchr/testify/require"
	q "go-ddd-book-manager/internal/application/user/query"
	"go-ddd-book-manager/internal/common/option"
	dc "go-ddd-book-manager/internal/domain/common"
	"go-ddd-book-manager/internal/domain/user"
	"go-ddd-book-manager/internal/infra/testutil"
	"testing"
	"time"
	"uuid"
)

func TestUserQuery(t *testing.T) {
	ctx := context.Background()
	i, _, _ := testutil.NewPostgres(t)
	repo := do.MustInvoke[user.Repository](i)
	query := do.MustInvoke[q.QueryService](i)
	id := uuid.NewV7()
	got, err := query.FindByID(ctx, id)
	require.NoError(t, err)
	require.Nil(t, got)
	audit := dc.RestoreEntityAudit(
		id,
		1,
		dc.RestoreAuditRecord(id, "Alice", time.Now().UTC().Truncate(time.Microsecond)),
		option.None[dc.AuditRecord](),
	)
	u, err := user.Restore(audit, "Alice", "alice@example.com", dc.UserRole)
	require.NoError(t, err)
	require.NoError(t, repo.Create(ctx, u))
	got, err = query.FindByID(ctx, id)
	require.NoError(t, err)
	require.Equal(t, id, got.ID)
	require.Equal(t, "alice@example.com", got.Email)
	require.Equal(t, dc.UserRole, got.Role)
}
