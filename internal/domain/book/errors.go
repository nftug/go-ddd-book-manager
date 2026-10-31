package book

import "go-ddd-book-manager/internal/domain/common"

var (
	OwnerUnchangedErr          = common.NewValidationError("book.owner", "unchanged")
	TitleEmptyErr              = common.NewValidationError("book.title", "empty")
	TitleTooLongErr            = common.NewValidationError("book.title", "too_long")
	ISBNLengthErr              = common.NewValidationError("book.isbn", "length")
	DescriptionTooLongErr      = common.NewValidationError("book.description", "too_long")
	AuthorEmptyErr             = common.NewValidationError("book.authors", "empty")
	AuthorTooLongErr           = common.NewValidationError("book.authors", "too_long")
	AuthorsEmptyErr            = common.NewValidationError("book.authors", "empty")
	AuthorsDuplicateErr        = common.NewValidationError("book.authors", "duplicate")
	CheckoutAlreadyActiveErr   = common.NewValidationError("book.checkout", "already_active")
	CheckoutNotActiveErr       = common.NewValidationError("book.checkout", "not_active")
	CheckoutAlreadyReturnedErr = common.NewValidationError("book.checkout", "already_returned")
)
