package book_test

import (
	"testing"
	"time"

	"go-ddd-book-manager/internal/common/option"
	"go-ddd-book-manager/internal/domain/book"
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

func newBook(t *testing.T, ctx common.DomainContext, owner user.Reference) book.Book {
	t.Helper()
	title, err := book.NewTitle("DDD")
	require.NoError(t, err)
	authors, err := book.NewAuthorList([]string{"Aki"})
	require.NoError(t, err)
	b, err := book.New(
		ctx,
		title,
		authors,
		option.None[book.ISBN](),
		option.None[book.Description](),
		owner,
	)
	require.NoError(t, err)
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

	require.Equal(t, "DDD", b.Title())
	authors := b.Authors()
	authors[0] = book.Author{}
	require.Equal(t, "Aki", b.Authors()[0].Name())
}

func TestOptionalISBNAndDescriptionAreExposedAsOptionalScalars(t *testing.T) {
	owner := user.RestoreReference(uuid.NewV4(), "Owner")
	ctx := common.NewDomainContext(
		common.NewActor(owner.ID(), owner.Name(), common.UserRole),
		testIDs{},
		testClock{time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
	)
	title, err := book.NewTitle("DDD")
	require.NoError(t, err)
	authors, err := book.NewAuthorList([]string{"Aki"})
	require.NoError(t, err)
	isbn, err := book.NewISBN("1234567890123")
	require.NoError(t, err)
	description, err := book.NewDescription("A domain-driven design book")
	require.NoError(t, err)
	b, err := book.New(ctx, title, authors, option.Some(isbn), option.Some(description), owner)
	require.NoError(t, err)

	assert.Equal(t, option.Some("1234567890123"), b.ISBN())
	assert.Equal(t, option.Some("A domain-driven design book"), b.Description())

	require.NoError(t, b.Update(
		ctx,
		title,
		authors,
		option.None[book.ISBN](),
		option.None[book.Description](),
	))

	assert.Equal(t, option.None[string](), b.ISBN())
	assert.Equal(t, option.None[string](), b.Description())
}

func TestRestoreMapsOptionalScalarsToValueObjects(t *testing.T) {
	owner := user.RestoreReference(uuid.NewV4(), "Owner")
	b := book.Restore(
		common.EntityAudit{},
		"DDD",
		[]book.Author{},
		option.Some("1234567890123"),
		option.Some("Restored description"),
		owner,
		[]book.Checkout{},
	)

	assert.Equal(t, option.Some("1234567890123"), b.ISBN())
	assert.Equal(t, option.Some("Restored description"), b.Description())
}
