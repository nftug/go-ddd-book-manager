package book

import (
	"fmt"

	"go-ddd-book-manager/internal/domain/common"
)

var (
	OwnerUnchangedErr          = common.NewValidationError("book.owner", "unchanged")
	OwnerNotFoundErr           = common.NewValidationError("book.owner_id", "not_found")
	TitleEmptyErr              = common.NewValidationError("book.title", "empty")
	TitleTooLongErr            = common.NewValidationError("book.title", "too_long")
	ISBNLengthErr              = common.NewValidationError("book.isbn", "length")
	DescriptionTooLongErr      = common.NewValidationError("book.description", "too_long")
	AuthorsEmptyErr            = common.NewValidationError("book.authors", "empty")
	AuthorsDuplicateErr        = common.NewValidationError("book.authors", "duplicate")
	CheckoutAlreadyActiveErr   = common.NewValidationError("book.checkout", "already_active")
	CheckoutNotActiveErr       = common.NewValidationError("book.checkout", "not_active")
	CheckoutAlreadyReturnedErr = common.NewValidationError("book.checkout", "already_returned")
)

func AuthorEmptyErr(index int) *common.ValidationError {
	return common.NewValidationError(fmt.Sprintf("book.author[%d]", index), "empty")
}

func AuthorTooLongErr(index int) *common.ValidationError {
	return common.NewValidationError(fmt.Sprintf("book.author[%d]", index), "too_long")
}
