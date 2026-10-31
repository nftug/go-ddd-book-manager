//go:build integration

package user_test

import (
	"context"
	"testing"
	"time"
	"uuid"

	"go-ddd-book-manager/internal/common/option"
	"go-ddd-book-manager/internal/domain/common"
	"go-ddd-book-manager/internal/domain/user"
	"go-ddd-book-manager/internal/infra/testutil"

	"github.com/samber/do/v2"

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
		1,
		common.RestoreAuditRecord(userID, "Original", stamp),
		option.None[common.AuditRecord](),
	)
	u, err := user.Restore(audit, "Original", "person@example.com", common.UserRole)
	require.NoError(t, err)
	updatedAudit := common.RestoreEntityAudit(
		userID,
		audit.Version(),
		audit.Created(),
		option.Some(common.RestoreAuditRecord(userID, "Original", stamp.Add(time.Second))),
	)
	updated, err := user.Restore(updatedAudit, "Changed", "changed@example.com", common.AdminRole)
	require.NoError(t, err)

	if !t.Run("find missing user", func(t *testing.T) {
		got, err := users.FindByID(ctx, uuid.NewV7())
		require.NoError(t, err)
		require.Nil(t, got)
	}) {
		return
	}
	if !t.Run("save and find user", func(t *testing.T) {
		require.NoError(t, users.Create(ctx, u))
		gotUser, err := users.FindByID(ctx, userID)
		require.NoError(t, err)
		require.Equal(t, u, *gotUser)
	}) {
		return
	}
	t.Run("update user", func(t *testing.T) {
		require.NoError(t, users.Update(ctx, updated))
		gotUser, err := users.FindByID(ctx, userID)
		require.NoError(t, err)
		require.Equal(t, userWithVersion(updated, 2), *gotUser)
	})
	t.Run("delete user", func(t *testing.T) {
		require.ErrorIs(t, users.Create(ctx, u), common.ErrConflict)
		require.ErrorIs(t, users.Update(ctx, updated), common.ErrConflict)
		require.ErrorIs(t, users.Delete(ctx, updated), common.ErrConflict)
		require.NoError(t, users.Delete(ctx, userWithVersion(updated, 2)))
		require.ErrorIs(t, users.Delete(ctx, userWithVersion(updated, 2)), common.ErrConflict)
		_, err := users.FindByID(ctx, userID)
		require.NoError(t, err)
	})
}

func userWithVersion(u user.User, version int64) user.User {
	audit := common.RestoreEntityAudit(u.ID(), version, u.Audit().Created(), u.Audit().Updated())
	restored, _ := user.Restore(audit, u.Name(), u.Email(), u.Role())
	return restored
}
