package book_test

import (
	"testing"
	"time"

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
	if err := b.Checkout(ctx); err != nil {
		t.Fatal(err)
	}
	if !b.IsCheckedOut() {
		t.Fatal("book should be checked out")
	}
	if err := b.Checkout(ctx); err == nil {
		t.Fatal("second active checkout should fail")
	}
	if err := b.Return(ctx); err != nil {
		t.Fatal(err)
	}
	if b.IsCheckedOut() {
		t.Fatal("book should no longer be checked out")
	}
	if err := b.Checkout(ctx); err != nil {
		t.Fatal(err)
	}
	if got := len(b.Checkouts()); got != 2 {
		t.Fatalf("got %d checkout records, want 2", got)
	}
	if b.Audit().Created() != created || b.Audit().Updated().IsSome() {
		t.Fatal("checkout flow must not update entity audit")
	}
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
	if err := b.Checkout(systemCtx); err != common.ErrForbidden {
		t.Fatalf("system checkout got %v, want forbidden", err)
	}
	if err := b.Checkout(ownerCtx); err != nil {
		t.Fatal(err)
	}
	if err := b.ValidateDeletion(ownerCtx.Actor()); err != nil {
		t.Fatalf("checked-out book should be deletable by owner: %v", err)
	}
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
	if checkout.CheckedOutAt().Location() != time.UTC {
		t.Fatal("checked-out timestamp should be normalized to UTC")
	}
	gotReturnedAt, ok := checkout.ReturnedAt().Get()
	if !ok || gotReturnedAt.Location() != time.UTC {
		t.Fatal("returned timestamp should be normalized to UTC")
	}
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

	assertValidationCode(t, b.Return(ctx), "book.checkout.not_active")
	if err := b.Checkout(ctx); err != nil {
		t.Fatal(err)
	}
	if err := b.Return(ctx); err != nil {
		t.Fatal(err)
	}
	assertValidationCode(t, b.Return(ctx), "book.checkout.already_returned")
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
	if err := b.Checkout(ownerCtx); err != nil {
		t.Fatal(err)
	}

	otherID := uuid.NewV4()
	otherCtx := common.NewDomainContext(
		common.NewActor(otherID, "Other", common.UserRole),
		testIDs{},
		testClock{now},
	)
	if err := b.Return(otherCtx); err != common.ErrForbidden {
		t.Fatalf("return by another user got %v, want forbidden", err)
	}
	if !b.IsCheckedOut() {
		t.Fatal("failed return should leave book checked out")
	}

	adminCtx := common.NewDomainContext(
		common.NewActor(uuid.NewV4(), "Admin", common.AdminRole),
		testIDs{},
		testClock{now.Add(time.Hour)},
	)
	if err := b.Return(adminCtx); err != nil {
		t.Fatalf("return by admin got %v, want nil", err)
	}
	if b.IsCheckedOut() {
		t.Fatal("successful admin return should close the checkout")
	}
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
	if list.IsCheckedOut() {
		t.Fatal("a returned latest checkout should make the list inactive")
	}

	input[1] = activeCheckout
	if list.IsCheckedOut() {
		t.Fatal("mutating restore input should not change the checkout list")
	}
	values := list.Values()
	values[1] = activeCheckout
	if list.IsCheckedOut() {
		t.Fatal("mutating returned values should not change the checkout list")
	}
}
