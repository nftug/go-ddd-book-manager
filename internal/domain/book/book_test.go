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

type testClock struct{ now time.Time }

func (c testClock) Now() time.Time { return c.now }

type testIDs struct{}

func (testIDs) NewUUIDv7() uuid.UUID { return uuid.NewV7() }

func newBook(t *testing.T, ctx common.DomainContext, owner user.Reference) book.Book {
	t.Helper()
	title, err := book.NewTitle("DDD")
	if err != nil {
		t.Fatal(err)
	}
	authors, err := book.NewAuthorList([]string{"Aki"})
	if err != nil {
		t.Fatal(err)
	}
	b, err := book.New(
		ctx,
		title,
		authors,
		option.None[book.ISBN](),
		option.None[book.Description](),
		owner,
	)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestAuthorsGetterReturnsCopy(t *testing.T) {
	owner := user.RestoreReference(uuid.NewV4(), "Owner")
	ctx := common.NewDomainContext(
		common.NewActor(owner.ID(), owner.Name(), common.UserRole),
		testIDs{},
		testClock{time.Now()},
	)
	b := newBook(t, ctx, owner)
	if got := b.Title(); got != "DDD" {
		t.Fatalf("Book.Title() = %q, want DDD", got)
	}
	authors := b.Authors()
	authors[0] = book.OrderedAuthor{}
	if b.Authors()[0].Name() != "Aki" {
		t.Fatal("modifying returned slice must not affect book")
	}
}

func TestOptionalISBNAndDescriptionAreExposedAsOptionalScalars(t *testing.T) {
	owner := user.RestoreReference(uuid.NewV4(), "Owner")
	ctx := common.NewDomainContext(
		common.NewActor(owner.ID(), owner.Name(), common.UserRole),
		testIDs{},
		testClock{time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
	)
	title, err := book.NewTitle("DDD")
	if err != nil {
		t.Fatal(err)
	}
	authors, err := book.NewAuthorList([]string{"Aki"})
	if err != nil {
		t.Fatal(err)
	}
	isbn, err := book.NewISBN("1234567890123")
	if err != nil {
		t.Fatal(err)
	}
	description, err := book.NewDescription("A domain-driven design book")
	if err != nil {
		t.Fatal(err)
	}
	b, err := book.New(ctx, title, authors, option.Some(isbn), option.Some(description), owner)
	if err != nil {
		t.Fatal(err)
	}

	if got, ok := b.ISBN().Get(); !ok || got != "1234567890123" {
		t.Fatalf("Book.ISBN() = (%q, %t), want (1234567890123, true)", got, ok)
	}
	if got, ok := b.Description().Get(); !ok || got != "A domain-driven design book" {
		t.Fatalf("Book.Description() = (%q, %t), want description and true", got, ok)
	}

	if err := b.Update(
		ctx,
		title,
		authors,
		option.None[book.ISBN](),
		option.None[book.Description](),
	); err != nil {
		t.Fatal(err)
	}
	if b.ISBN().IsSome() || b.Description().IsSome() {
		t.Fatal("clearing optional values should expose None from getters")
	}
}

func TestRestoreMapsOptionalScalarsToValueObjects(t *testing.T) {
	owner := user.RestoreReference(uuid.NewV4(), "Owner")
	b := book.Restore(
		common.EntityAudit{},
		"DDD",
		[]book.OrderedAuthor{},
		option.Some("1234567890123"),
		option.Some("Restored description"),
		owner,
		[]book.Checkout{},
	)
	if got, ok := b.ISBN().Get(); !ok || got != "1234567890123" {
		t.Fatalf("restored ISBN = (%q, %t), want scalar value and true", got, ok)
	}
	if got, ok := b.Description().Get(); !ok || got != "Restored description" {
		t.Fatalf("restored description = (%q, %t), want scalar value and true", got, ok)
	}
}
