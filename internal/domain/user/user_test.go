package user_test

import (
	"testing"
	"time"

	"go-ddd-book-manager/internal/domain/common"
	"go-ddd-book-manager/internal/domain/user"
	"uuid"
)

type testClock struct{ now time.Time }

func (c testClock) Now() time.Time { return c.now }

type testIDs struct{}

func (testIDs) NewUUIDv7() uuid.UUID { return uuid.NewV7() }

func TestSyncOnlyUpdatesAuditWhenProfileChanges(t *testing.T) {
	id := uuid.NewV4()
	name, _ := user.NewName("Aki")
	email, _ := user.NewEmail("aki@example.com")
	ctx := common.NewDomainContext(
		common.NewActor(id, "Aki", common.UserRole),
		testIDs{},
		testClock{time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
	)
	u, err := user.New(ctx, id, name, email, common.UserRole)
	if err != nil {
		t.Fatal(err)
	}
	if u.Name() != name.Value() || u.Email() != email.Value() {
		t.Fatal("user getters should return scalar name and email values")
	}
	if err := u.Sync(ctx, name, email, common.UserRole); err != nil {
		t.Fatal(err)
	}
	if u.Audit().Updated().IsSome() {
		t.Fatal("unchanged profile must not update audit")
	}

	changedName, _ := user.NewName("Aki Tanaka")
	if err := u.Sync(ctx, changedName, email, common.UserRole); err != nil {
		t.Fatal(err)
	}
	if u.Name() != changedName.Value() || u.Audit().Updated().IsNone() {
		t.Fatal("changed profile must update user and audit")
	}
}

func TestSyncRejectsDifferentUser(t *testing.T) {
	id := uuid.NewV4()
	name, _ := user.NewName("Aki")
	email, _ := user.NewEmail("aki@example.com")
	ctx := common.NewDomainContext(
		common.NewActor(id, "Aki", common.UserRole),
		testIDs{},
		testClock{time.Now()},
	)
	u, err := user.New(ctx, uuid.NewV4(), name, email, common.UserRole)
	if err != nil {
		t.Fatal(err)
	}
	if err := u.Sync(ctx, name, email, common.AdminRole); err != common.ErrForbidden {
		t.Fatalf("got %v, want forbidden", err)
	}
	if u.Role() != common.UserRole || u.Audit().Updated().IsSome() {
		t.Fatal("rejected sync must leave user unchanged")
	}
}

func TestSystemActorUsesZeroUUIDAndCanSync(t *testing.T) {
	name, _ := user.NewName("Aki")
	email, _ := user.NewEmail("aki@example.com")
	ctx := common.NewDomainContext(common.NewSystemActor(), testIDs{}, testClock{time.Now()})
	u, err := user.New(ctx, uuid.NewV4(), name, email, common.UserRole)
	if err != nil {
		t.Fatal(err)
	}
	if ctx.Actor().ID() != uuid.Nil() || ctx.Actor().UserName() != "System" {
		t.Fatal("system audit identity must use zero UUID and System name")
	}
	changed, _ := user.NewName("Aki Updated")
	if err := u.Sync(ctx, changed, email, common.AdminRole); err != nil {
		t.Fatal(err)
	}
	if u.Name() != changed.Value() || u.Role() != common.AdminRole {
		t.Fatal("system sync should apply authoritative profile")
	}
	updated, ok := u.Audit().Updated().Get()
	if !ok || updated.UserID() != uuid.Nil() || updated.UserName() != "System" {
		t.Fatal("sync audit should identify System")
	}
}
