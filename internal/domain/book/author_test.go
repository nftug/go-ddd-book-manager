package book_test

import (
	"strings"
	"testing"

	"go-ddd-book-manager/internal/domain/book"
)

func TestAuthorNameConstructorRestoreAndValidation(t *testing.T) {
	name, err := book.NewAuthorName("Aki")
	if err != nil {
		t.Fatal(err)
	}
	if got := name.Value(); got != "Aki" {
		t.Fatalf("NewAuthorName value = %q, want Aki", got)
	}
	if got := book.RestoreAuthorName("Restored").Value(); got != "Restored" {
		t.Fatalf("RestoreAuthorName value = %q, want Restored", got)
	}

	for _, tc := range []struct {
		name  string
		value string
		want  string
	}{
		{name: "empty", value: "", want: "book.author.empty"},
		{name: "too long", value: strings.Repeat("あ", 256), want: "book.author.too_long"},
	} {
		t.Run(tc.want, func(t *testing.T) {
			_, err := book.NewAuthorName(tc.value)
			assertValidationCode(t, err, tc.want)
		})
	}
}

func TestOrderedAuthorRestorePreservesNameAndOrder(t *testing.T) {
	name := book.RestoreAuthorName("Aki")
	author := book.RestoreOrderedAuthor(name, 3)
	if author.Name() != name.Value() {
		t.Fatalf("OrderedAuthor name = %q, want %q", author.Name(), name.Value())
	}
	if got := author.OrderIndex(); got != 3 {
		t.Fatalf("OrderedAuthor order = %d, want 3", got)
	}
}

func TestAuthorListConstructorOrderingAndValidation(t *testing.T) {
	list, err := book.NewAuthorList([]string{"Aki", "Mika"})
	if err != nil {
		t.Fatal(err)
	}
	values := list.Values()
	if len(values) != 2 {
		t.Fatalf("AuthorList length = %d, want 2", len(values))
	}
	for i, want := range []string{"Aki", "Mika"} {
		if got := values[i].Name(); got != want {
			t.Errorf("author %d name = %q, want %q", i, got, want)
		}
		if got := values[i].OrderIndex(); got != i {
			t.Errorf("author %d order = %d, want %d", i, got, i)
		}
	}

	for _, tc := range []struct {
		name  string
		input []string
		want  string
	}{
		{name: "empty list", input: nil, want: "book.authors.empty"},
		{name: "duplicate names", input: []string{"Aki", "Aki"}, want: "book.authors.duplicate"},
		{name: "invalid name", input: []string{"Aki", ""}, want: "book.author.empty"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := book.NewAuthorList(tc.input)
			assertValidationCode(t, err, tc.want)
		})
	}
}

func TestRestoreAuthorListCopiesInputAndValues(t *testing.T) {
	first := book.RestoreOrderedAuthor(book.RestoreAuthorName("Aki"), 0)
	input := []book.OrderedAuthor{first}
	list := book.RestoreAuthorList(input)
	input[0] = book.OrderedAuthor{}
	if got := list.Values()[0].Name(); got != "Aki" {
		t.Fatalf("mutating restore input changed list value to %q", got)
	}

	values := list.Values()
	values[0] = book.OrderedAuthor{}
	if got := list.Values()[0].Name(); got != "Aki" {
		t.Fatalf("mutating returned values changed list value to %q", got)
	}
}
