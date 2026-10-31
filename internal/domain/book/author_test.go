package book_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go-ddd-book-manager/internal/domain/book"
)

func TestAuthorNameConstructorRestoreAndValidation(t *testing.T) {
	name, err := book.NewAuthorName("Aki")
	require.NoError(t, err)

	require.Equal(t, "Aki", name.Value())
	require.Equal(t, "Restored", book.RestoreAuthorName("Restored").Value())

	for _, tc := range []struct {
		name  string
		value string
		want  string
	}{
		{name: "empty", value: "", want: "empty"},
		{name: "too long", value: strings.Repeat("あ", 256), want: "too_long"},
	} {
		t.Run(tc.want, func(t *testing.T) {
			_, err := book.NewAuthorName(tc.value)
			requireValidationError(t, err, "book.authors", tc.want)
		})
	}
}

func TestOrderedAuthorRestorePreservesNameAndOrder(t *testing.T) {
	name := book.RestoreAuthorName("Aki")
	author := book.RestoreOrderedAuthor(name, 3)

	assert.Equal(t, name.Value(), author.Name())
	assert.Equal(t, 3, author.OrderIndex())
}

func TestAuthorListConstructorOrderingAndValidation(t *testing.T) {
	list, err := book.NewAuthorList([]string{"Aki", "Mika"})
	require.NoError(t, err)

	values := list.Values()
	require.Len(t, values, 2)
	for i, want := range []string{"Aki", "Mika"} {
		assert.Equal(t, want, values[i].Name())
		assert.Equal(t, i, values[i].OrderIndex())
	}

	for _, tc := range []struct {
		name  string
		input []string
		want  string
	}{
		{name: "empty list", input: nil, want: "empty"},
		{name: "duplicate names", input: []string{"Aki", "Aki"}, want: "duplicate"},
		{name: "invalid name", input: []string{"Aki", ""}, want: "empty"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := book.NewAuthorList(tc.input)
			requireValidationError(t, err, "book.authors", tc.want)
		})
	}
}

func TestRestoreAuthorListCopiesInputAndValues(t *testing.T) {
	first := book.RestoreOrderedAuthor(book.RestoreAuthorName("Aki"), 0)
	input := []book.OrderedAuthor{first}
	list := book.RestoreAuthorList(input)

	input[0] = book.OrderedAuthor{}
	require.Equal(t, "Aki", list.Values()[0].Name())

	values := list.Values()
	values[0] = book.OrderedAuthor{}
	require.Equal(t, "Aki", list.Values()[0].Name())
}
