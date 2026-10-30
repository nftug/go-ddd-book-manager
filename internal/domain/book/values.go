package book

import (
	"unicode/utf8"

	"go-ddd-book-manager/internal/domain/common"
)

type Title struct{ value string }

func NewTitle(value string) (Title, error) {
	if value == "" {
		return Title{}, common.NewValidationError("book.title.empty")
	}
	if utf8.RuneCountInString(value) > 255 {
		return Title{}, common.NewValidationError("book.title.too_long")
	}
	return Title{value}, nil
}
func RestoreTitle(value string) Title { return Title{value} }
func (v Title) Value() string         { return v.value }

type ISBN struct{ value string }

func NewISBN(value string) (ISBN, error) {
	if len(value) != 13 {
		return ISBN{}, common.NewValidationError("book.isbn.length")
	}
	return ISBN{value: value}, nil
}
func RestoreISBN(value string) ISBN { return ISBN{value: value} }
func (i ISBN) Value() string        { return i.value }

type Description struct{ value string }

func NewDescription(value string) (Description, error) {
	if utf8.RuneCountInString(value) > 1000 {
		return Description{}, common.NewValidationError("book.description.too_long")
	}
	return Description{value: value}, nil
}
func RestoreDescription(value string) Description { return Description{value: value} }
func (d Description) Value() string               { return d.value }
