//go:build integration

package user_test

import (
	"context"
	"testing"
	"time"
	"uuid"

	"github.com/samber/do/v2"
	"go-ddd-book-manager/internal/common/option"
	"go-ddd-book-manager/internal/domain/common"
	"go-ddd-book-manager/internal/domain/user"
	"go-ddd-book-manager/internal/infra/testutil"

	"github.com/stretchr/testify/require"
)

func TestRepository(t *testing.T) {
	ctx := context.Background()
	i, _, _ := testutil.NewPostgres(t)
	users := do.MustInvoke[user.Repository](i)

	stamp := time.Now().UTC().Truncate(time.Microsecond)
	userID := uuid.NewV7()
	audit := common.RestoreEntityAudit(
		userID,
		common.RestoreAuditRecord(userID, "Original", stamp),
		option.None[common.AuditRecord](),
	)
	u, err := user.Restore(audit, "Original", "person@example.com", common.UserRole)
	require.NoError(t, err)
	updatedAudit := common.RestoreEntityAudit(
		userID,
		audit.Created(),
		option.Some(common.RestoreAuditRecord(userID, "Original", stamp.Add(time.Second))),
	)
	updated, err := user.Restore(updatedAudit, "Changed", "changed@example.com", common.AdminRole)
	require.NoError(t, err)

	if !t.Run("find missing user", func(t *testing.T) {
		_, err := users.FindByID(ctx, uuid.NewV7())
		require.ErrorIs(t, err, common.ErrNotFound)
	}) {
		return
	}
	if !t.Run("save and find user", func(t *testing.T) {
		require.NoError(t, users.Save(ctx, u))
		gotUser, err := users.FindByID(ctx, userID)
		require.NoError(t, err)
		require.Equal(t, u, *gotUser)
	}) {
		return
	}
	t.Run("update user", func(t *testing.T) {
		require.NoError(t, users.Save(ctx, updated))
		gotUser, err := users.FindByID(ctx, userID)
		require.NoError(t, err)
		require.Equal(t, updated, *gotUser)
	})
}
