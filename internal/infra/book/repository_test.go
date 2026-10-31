//go:build integration

package book_test

import (
	"context"
	"testing"
	"time"
	"uuid"

	"go-ddd-book-manager/internal/common/option"
	"go-ddd-book-manager/internal/domain/book"
	"go-ddd-book-manager/internal/domain/common"
	"go-ddd-book-manager/internal/domain/user"
	"go-ddd-book-manager/internal/infra/testutil"

	"github.com/samber/do/v2"

	"github.com/stretchr/testify/require"
)

func TestRepository(t *testing.T) {
	ctx := context.Background()
	i, _, _ := testutil.NewPostgres(t)
	books := do.MustInvoke[book.Repository](i)

	userID := uuid.NewV7()
	stamp := time.Now().UTC().Truncate(time.Microsecond)
	audit := common.RestoreEntityAudit(
		userID,
		common.RestoreAuditRecord(userID, "Original", stamp),
		option.None[common.AuditRecord](),
	)
	bookID := uuid.NewV7()
	checkoutID := uuid.NewV7()
	ba := common.RestoreEntityAudit(bookID, audit.Created(), option.None[common.AuditRecord]())
	ref := user.RestoreReference(userID, "Snapshot")
	checkout := book.RestoreCheckout(checkoutID, ref, stamp, option.None[time.Time]())
	b := book.Restore(
		ba,
		"Title",
		[]book.Author{book.RestoreAuthor("First", 0), book.RestoreAuthor("Second", 1)},
		option.Some("1234567890123"),
		option.None[string](),
		ref,
		[]book.Checkout{checkout},
	)
	if !t.Run("find missing book", func(t *testing.T) {
		got, err := books.FindByID(ctx, uuid.NewV7())
		require.NoError(t, err)
		require.Nil(t, got)
	}) {
		return
	}
	if !t.Run("save and find book", func(t *testing.T) {
		require.NoError(t, books.Save(ctx, b))
		got, err := books.FindByID(ctx, bookID)
		require.NoError(t, err)
		require.Equal(t, b, *got)
	}) {
		return
	}

	returned := book.RestoreCheckout(checkoutID, ref, stamp, option.Some(stamp.Add(time.Hour)))
	b = book.Restore(
		ba,
		"Updated",
		[]book.Author{book.RestoreAuthor("New", 0)},
		option.None[string](),
		option.Some("description"),
		ref,
		[]book.Checkout{returned},
	)
	if !t.Run("update book", func(t *testing.T) {
		require.NoError(t, books.Save(ctx, b))
		got, err := books.FindByID(ctx, bookID)
		require.NoError(t, err)
		require.Equal(t, b, *got)
	}) {
		return
	}
	if !t.Run("rollback invalid save", func(t *testing.T) {
		invalid := book.Restore(
			ba,
			"Should roll back",
			[]book.Author{book.RestoreAuthor("Duplicate", 0), book.RestoreAuthor("Duplicate", 1)},
			option.None[string](),
			option.None[string](),
			ref,
			nil,
		)
		require.Error(t, books.Save(ctx, invalid))
		got, err := books.FindByID(ctx, bookID)
		require.NoError(t, err)
		require.Equal(t, b, *got)
	}) {
		return
	}
	t.Run("delete book", func(t *testing.T) {
		require.NoError(t, books.Delete(ctx, bookID))
		require.NoError(t, books.Delete(ctx, bookID))
		_, err := books.FindByID(ctx, bookID)
		require.NoError(t, err)
	})
}
