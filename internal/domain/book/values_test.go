package book_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go-ddd-book-manager/internal/domain/book"
)

func TestScalarValueConstructorsAndRestoreFunctions(t *testing.T) {
	title, err := book.NewTitle("DDD")
	require.NoError(t, err)

	require.Equal(t, "DDD", title.Value())
	require.Equal(t, "Restored", book.RestoreTitle("Restored").Value())

	isbn, err := book.NewISBN("1234567890123")
	require.NoError(t, err)

	require.Equal(t, "1234567890123", isbn.Value())
	require.Equal(t, "Restored ISBN", book.RestoreISBN("Restored ISBN").Value())

	description, err := book.NewDescription("A description")
	require.NoError(t, err)

	require.Equal(t, "A description", description.Value())
	require.Equal(
		t,
		"Restored description",
		book.RestoreDescription("Restored description").Value(),
	)
}

func TestScalarValueValidationCodesAndBoundaries(t *testing.T) {
	_, err := book.NewTitle("")
	requireValidationError(t, err, "book.title", "empty")

	_, err = book.NewTitle(strings.Repeat("a", 255))
	require.NoError(t, err)

	_, err = book.NewTitle(strings.Repeat("あ", 256))
	requireValidationError(t, err, "book.title", "too_long")

	_, err = book.NewISBN("123")
	requireValidationError(t, err, "book.isbn", "length")

	_, err = book.NewDescription(strings.Repeat("あ", 1000))
	require.NoError(t, err)

	_, err = book.NewDescription(strings.Repeat("あ", 1001))
	requireValidationError(t, err, "book.description", "too_long")
}
