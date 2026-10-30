package book_test

import (
	"strings"
	"testing"

	"go-ddd-book-manager/internal/domain/book"
)

func TestScalarValueConstructorsAndRestoreFunctions(t *testing.T) {
	title, err := book.NewTitle("DDD")
	if err != nil || title.Value() != "DDD" {
		t.Fatalf("NewTitle got (%q, %v), want (DDD, nil)", title.Value(), err)
	}
	if got := book.RestoreTitle("Restored").Value(); got != "Restored" {
		t.Fatalf("RestoreTitle got %q, want Restored", got)
	}

	isbn, err := book.NewISBN("1234567890123")
	if err != nil || isbn.Value() != "1234567890123" {
		t.Fatalf("NewISBN got (%q, %v), want (1234567890123, nil)", isbn.Value(), err)
	}
	if got := book.RestoreISBN("Restored ISBN").Value(); got != "Restored ISBN" {
		t.Fatalf("RestoreISBN got %q, want Restored ISBN", got)
	}

	description, err := book.NewDescription("A description")
	if err != nil || description.Value() != "A description" {
		t.Fatalf("NewDescription got (%q, %v), want (A description, nil)", description.Value(), err)
	}
	if got := book.RestoreDescription("Restored description").Value(); got != "Restored description" {
		t.Fatalf("RestoreDescription got %q, want Restored description", got)
	}
}

func TestScalarValueValidationCodesAndBoundaries(t *testing.T) {
	_, err := book.NewTitle("")
	assertValidationCode(t, err, "book.title.empty")
	if _, err := book.NewTitle(strings.Repeat("a", 255)); err != nil {
		t.Fatalf("title with 255 runes should be valid: %v", err)
	}
	_, err = book.NewTitle(strings.Repeat("あ", 256))
	assertValidationCode(t, err, "book.title.too_long")

	_, err = book.NewISBN("123")
	assertValidationCode(t, err, "book.isbn.length")

	if _, err := book.NewDescription(strings.Repeat("あ", 1000)); err != nil {
		t.Fatalf("description with 1000 runes should be valid: %v", err)
	}
	_, err = book.NewDescription(strings.Repeat("あ", 1001))
	assertValidationCode(t, err, "book.description.too_long")
}
