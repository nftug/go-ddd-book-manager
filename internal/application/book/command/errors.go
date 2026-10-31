package command

import "go-ddd-book-manager/internal/domain/common"

var OwnerNotFoundErr = common.NewValidationError("book.owner_id", "not_found")
