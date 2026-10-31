package user_test

import (
	"testing"
	"time"

	"go-ddd-book-manager/internal/domain/common"
	"go-ddd-book-manager/internal/domain/user"
	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testClock struct{ now time.Time }

func (c testClock) Now() time.Time { return c.now }

type testIDs struct{}

func (testIDs) NewUUIDv7() uuid.UUID { return uuid.NewV7() }

func newSystemCreatedUser(
	t *testing.T,
	id uuid.UUID,
	name user.Name,
	email user.Email,
	role common.ActorRole,
) user.User {
	t.Helper()
	ctx := common.NewDomainContext(common.NewSystemActor(), testIDs{}, testClock{time.Now()})
	u, err := user.New(ctx, id, name.Value(), email.Value(), role)
	require.NoError(t, err)
	return u
}

func TestSyncOnlyUpdatesAuditWhenProfileChanges(t *testing.T) {
	id := uuid.NewV4()
	name, err := user.NewName("Aki")
	require.NoError(t, err)
	email, err := user.NewEmail("aki@example.com")
	require.NoError(t, err)
	ctx := common.NewDomainContext(
		common.NewSystemActor(),
		testIDs{},
		testClock{time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
	)
	u := newSystemCreatedUser(t, id, name, email, common.UserRole)

	changed, err := u.Sync(ctx, name.Value(), email.Value(), common.UserRole)
	require.NoError(t, err)
	assert.False(t, changed)
	assert.True(t, u.Audit().Updated().IsNone())

	changedName, err := user.NewName("Aki Tanaka")
	require.NoError(t, err)

	changed, err = u.Sync(ctx, changedName.Value(), email.Value(), common.UserRole)
	require.NoError(t, err)
	assert.True(t, changed)

	assert.Equal(t, changedName.Value(), u.Name())
	assert.True(t, u.Audit().Updated().IsSome())
}

func TestSyncRejectsNonSystemActor(t *testing.T) {
	id := uuid.NewV4()
	name, err := user.NewName("Aki")
	require.NoError(t, err)
	email, err := user.NewEmail("aki@example.com")
	require.NoError(t, err)
	ctx := common.NewDomainContext(
		common.NewActor(id, "Aki", common.AdminRole),
		testIDs{},
		testClock{time.Now()},
	)

	u := newSystemCreatedUser(t, uuid.NewV4(), name, email, common.UserRole)

	_, err = u.Sync(ctx, name.Value(), email.Value(), common.AdminRole)
	require.ErrorIs(t, err, common.ErrForbidden)

	assert.Equal(t, common.UserRole, u.Role())
	assert.True(t, u.Audit().Updated().IsNone())
}

func TestSyncRejectsInvalidAndSystemRoles(t *testing.T) {
	name, err := user.NewName("Aki")
	require.NoError(t, err)
	email, err := user.NewEmail("aki@example.com")
	require.NoError(t, err)
	ctx := common.NewDomainContext(common.NewSystemActor(), testIDs{}, testClock{time.Now()})
	u := newSystemCreatedUser(t, uuid.NewV4(), name, email, common.UserRole)

	_, err = u.Sync(ctx, name.Value(), email.Value(), common.ActorRole("invalid"))
	require.ErrorIs(t, err, user.RoleInvalidErr)
	_, err = u.Sync(ctx, name.Value(), email.Value(), common.SystemRole)
	require.ErrorIs(t, err, user.RoleInvalidErr)
	assert.Equal(t, common.UserRole, u.Role())
	assert.True(t, u.Audit().Updated().IsNone())
}

func TestSystemActorUsesZeroUUIDAndCanSync(t *testing.T) {
	name, err := user.NewName("Aki")
	require.NoError(t, err)
	email, err := user.NewEmail("aki@example.com")
	require.NoError(t, err)
	ctx := common.NewDomainContext(common.NewSystemActor(), testIDs{}, testClock{time.Now()})

	u := newSystemCreatedUser(t, uuid.NewV4(), name, email, common.UserRole)

	assert.Equal(t, uuid.Nil(), ctx.Actor().ID())
	assert.Equal(t, "System", ctx.Actor().UserName())
	assert.Equal(t, common.SystemRole, ctx.Actor().Role())
	assert.True(t, ctx.Actor().IsSystem())
	assert.False(t, ctx.Actor().IsAdmin())

	changed, err := user.NewName("Aki Updated")
	require.NoError(t, err)
	profileChanged, err := u.Sync(ctx, changed.Value(), email.Value(), common.AdminRole)
	require.NoError(t, err)
	assert.True(t, profileChanged)

	assert.Equal(t, changed.Value(), u.Name())
	assert.Equal(t, common.AdminRole, u.Role())

	updated, ok := u.Audit().Updated().Get()
	require.True(t, ok)

	assert.Equal(t, uuid.Nil(), updated.UserID())
	assert.Equal(t, "System", updated.UserName())
}

func TestNewRejectsSystemRole(t *testing.T) {
	name, err := user.NewName("Aki")
	require.NoError(t, err)
	email, err := user.NewEmail("aki@example.com")
	require.NoError(t, err)
	ctx := common.NewDomainContext(common.NewSystemActor(), testIDs{}, testClock{time.Now()})

	_, err = user.New(ctx, uuid.NewV4(), name.Value(), email.Value(), common.SystemRole)
	require.ErrorIs(t, err, user.RoleInvalidErr)
}

func TestNewRequiresSystemActor(t *testing.T) {
	name, err := user.NewName("Aki")
	require.NoError(t, err)
	email, err := user.NewEmail("aki@example.com")
	require.NoError(t, err)
	ctx := common.NewDomainContext(
		common.NewActor(uuid.NewV4(), "Aki", common.AdminRole),
		testIDs{},
		testClock{time.Now()},
	)

	_, err = user.New(ctx, uuid.NewV4(), name.Value(), email.Value(), common.UserRole)
	require.ErrorIs(t, err, common.ErrForbidden)
}

func TestNewValidatesNameAndEmail(t *testing.T) {
	ctx := common.NewDomainContext(common.NewSystemActor(), testIDs{}, testClock{time.Now()})
	tests := []struct{ name, inputName, email, field string }{
		{name: "empty name", inputName: "", email: "aki@example.com", field: "user.name"},
		{name: "invalid email", inputName: "Aki", email: "invalid", field: "user.email"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := user.New(ctx, uuid.NewV4(), tt.inputName, tt.email, common.UserRole)
			var validation *common.ValidationError
			require.ErrorAs(t, err, &validation)
			require.Equal(t, tt.field, validation.Field())
		})
	}
}
