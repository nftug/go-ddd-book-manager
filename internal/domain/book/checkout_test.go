package book_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go-ddd-book-manager/internal/common/option"
	"go-ddd-book-manager/internal/domain/book"
	"go-ddd-book-manager/internal/domain/common"
	"go-ddd-book-manager/internal/domain/user"
	"uuid"
)

func TestCheckoutReturnAndRecheckoutDoNotUpdateEntityAudit(t *testing.T) {
	userID := uuid.NewV4()
	actor := common.NewActor(userID, "Aki", common.UserRole)
	ctx := common.NewDomainContext(
		actor,
		testIDs{},
		testClock{time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
	)
	owner := user.RestoreReference(userID, "Aki")
	b := newBook(t, ctx, owner)
	created := b.Audit().Created()

	require.NoError(t, b.Checkout(ctx))
	require.True(t, b.IsCheckedOut())

	requireValidationError(t, b.Checkout(ctx), "book.checkout", "already_active")
	require.NoError(t, b.Return(ctx))
	require.False(t, b.IsCheckedOut())

	require.NoError(t, b.Checkout(ctx))

	require.Len(t, b.Checkouts(), 2)
	require.Equal(t, created, b.Audit().Created())
	require.True(t, b.Audit().Updated().IsNone())
}

func TestCheckoutRejectsSystemAndDeleteAllowsCheckedOutBook(t *testing.T) {
	owner := user.RestoreReference(uuid.NewV4(), "Owner")
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	ownerCtx := common.NewDomainContext(
		common.NewActor(owner.ID(), owner.Name(), common.UserRole),
		testIDs{},
		testClock{now},
	)
	b := newBook(t, ownerCtx, owner)
	systemCtx := common.NewDomainContext(common.NewSystemActor(), testIDs{}, testClock{now})

	require.ErrorIs(t, b.Checkout(systemCtx), common.ErrForbidden)

	require.NoError(t, b.Checkout(ownerCtx))
	require.NoError(t, b.ValidateDeletion(ownerCtx.Actor()))
}

func TestRestoreCheckoutNormalizesTimestamps(t *testing.T) {
	checkedOutAt := time.Date(2026, 1, 1, 9, 0, 0, 0, time.FixedZone("JST", 9*60*60))
	returnedAt := checkedOutAt.Add(time.Hour)
	checkout := book.RestoreCheckout(
		uuid.NewV4(),
		user.RestoreReference(uuid.NewV4(), "Aki"),
		checkedOutAt,
		option.Some(returnedAt),
	)

	require.Equal(t, time.UTC, checkout.CheckedOutAt().Location())
	gotReturnedAt, ok := checkout.ReturnedAt().Get()
	require.True(t, ok)
	require.Equal(t, time.UTC, gotReturnedAt.Location())
}

func TestReturnRequiresAnActiveCheckout(t *testing.T) {
	owner := user.RestoreReference(uuid.NewV4(), "Owner")
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	ctx := common.NewDomainContext(
		common.NewActor(owner.ID(), owner.Name(), common.UserRole),
		testIDs{},
		testClock{now},
	)
	b := newBook(t, ctx, owner)

	requireValidationError(t, b.Return(ctx), "book.checkout", "not_active")

	require.NoError(t, b.Checkout(ctx))
	require.NoError(t, b.Return(ctx))

	requireValidationError(t, b.Return(ctx), "book.checkout", "already_returned")
}

func TestReturnAllowsOnlyCheckoutOwnerOrAdmin(t *testing.T) {
	owner := user.RestoreReference(uuid.NewV4(), "Owner")
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	ownerCtx := common.NewDomainContext(
		common.NewActor(owner.ID(), owner.Name(), common.UserRole),
		testIDs{},
		testClock{now},
	)
	b := newBook(t, ownerCtx, owner)

	require.NoError(t, b.Checkout(ownerCtx))

	otherID := uuid.NewV4()
	otherCtx := common.NewDomainContext(
		common.NewActor(otherID, "Other", common.UserRole),
		testIDs{},
		testClock{now},
	)
	require.ErrorIs(t, b.Return(otherCtx), common.ErrForbidden)
	require.True(t, b.IsCheckedOut())

	adminCtx := common.NewDomainContext(
		common.NewActor(uuid.NewV4(), "Admin", common.AdminRole),
		testIDs{},
		testClock{now.Add(time.Hour)},
	)
	require.NoError(t, b.Return(adminCtx))
	require.False(t, b.IsCheckedOut())
}

func TestRestoreCheckoutListUsesLatestCheckoutAndReturnsCopies(t *testing.T) {
	checkedOutAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	owner := user.RestoreReference(uuid.NewV4(), "Owner")
	activeCheckout := book.RestoreCheckout(
		uuid.NewV4(),
		owner,
		checkedOutAt,
		option.None[time.Time](),
	)
	returnedAt := checkedOutAt.Add(2 * time.Hour)
	latestCheckout := book.RestoreCheckout(
		uuid.NewV4(), owner, checkedOutAt.Add(time.Hour), option.Some(returnedAt),
	)
	input := []book.Checkout{activeCheckout, latestCheckout}
	list := book.RestoreCheckoutList(input)
	require.False(t, list.IsCheckedOut())

	input[1] = activeCheckout
	require.False(t, list.IsCheckedOut())
	values := list.Values()
	values[1] = activeCheckout
	require.False(t, list.IsCheckedOut())
}
