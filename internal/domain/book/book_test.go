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
	b, err := book.New(
		ctx,
		"DDD",
		[]string{"Aki"},
		option.None[string](),
		option.None[string](),
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
	b, err := book.New(
		ctx,
		"DDD",
		[]string{"Aki"},
		option.Some("1234567890123"),
		option.Some("A domain-driven design book"),
		owner,
	)
	require.NoError(t, err)

	assert.Equal(t, option.Some("1234567890123"), b.ISBN())
	assert.Equal(t, option.Some("A domain-driven design book"), b.Description())

	require.NoError(t, b.Update(
		ctx,
		"DDD",
		[]string{"Aki"},
		option.None[string](),
		option.None[string](),
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

func TestNewValidatesRawBookValues(t *testing.T) {
	owner := user.RestoreReference(uuid.NewV4(), "Owner")
	ctx := common.NewDomainContext(
		common.NewActor(owner.ID(), owner.Name(), common.UserRole),
		testIDs{},
		testClock{time.Now()},
	)
	tests := []struct {
		name        string
		title       string
		authors     []string
		isbn        option.Option[string]
		description option.Option[string]
		field       string
	}{
		{name: "title", title: "", authors: []string{"Aki"}, field: "book.title"},
		{name: "author", title: "DDD", authors: []string{""}, field: "book.author[0]"},
		{
			name:    "ISBN",
			title:   "DDD",
			authors: []string{"Aki"},
			isbn:    option.Some("bad"),
			field:   "book.isbn",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := book.New(ctx, tt.title, tt.authors, tt.isbn, tt.description, owner)
			var validation *common.ValidationError
			require.ErrorAs(t, err, &validation)
			require.Equal(t, tt.field, validation.Field())
		})
	}
}

func TestUpdateInvalidValuesLeaveBookAndAuditUnchanged(t *testing.T) {
	owner := user.RestoreReference(uuid.NewV4(), "Owner")
	ctx := common.NewDomainContext(
		common.NewActor(owner.ID(), owner.Name(), common.UserRole),
		testIDs{},
		testClock{time.Now()},
	)
	b := newBook(t, ctx, owner)
	before := b.Audit()
	err := b.Update(ctx, "Updated", []string{"Aki"}, option.Some("invalid"), option.None[string]())
	var validation *common.ValidationError
	require.ErrorAs(t, err, &validation)
	require.Equal(t, "DDD", b.Title())
	require.Equal(t, before, b.Audit())
	require.True(t, b.ISBN().IsNone())
}
