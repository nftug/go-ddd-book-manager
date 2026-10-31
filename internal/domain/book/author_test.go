package book_test

import (
	"fmt"
	"strings"
	"testing"

	"go-ddd-book-manager/internal/domain/book"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrderedAuthorRestorePreservesNameAndOrder(t *testing.T) {
	name := "Aki"
	author := book.RestoreAuthor(name, 3)

	assert.Equal(t, name, author.Name())
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
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := book.NewAuthorList(tc.input)
			requireValidationError(t, err, "book.authors", tc.want)
		})
	}

	for _, index := range []int{0, 1, 2} {
		names := []string{"Aki", "Mika", "Yui"}
		names[index] = ""
		_, err := book.NewAuthorList(names)
		requireValidationError(t, err, fmt.Sprintf("book.author[%d]", index), "empty")
	}

	_, err = book.NewAuthorList([]string{"Aki", strings.Repeat("あ", 256)})
	requireValidationError(t, err, "book.author[1]", "too_long")
}

func TestRestoreAuthorListCopiesInputAndValues(t *testing.T) {
	first := book.RestoreAuthor("Aki", 0)
	input := []book.Author{first}
	list := book.RestoreAuthorList(input)

	input[0] = book.Author{}
	require.Equal(t, "Aki", list.Values()[0].Name())

	values := list.Values()
	values[0] = book.Author{}
	require.Equal(t, "Aki", list.Values()[0].Name())
}
