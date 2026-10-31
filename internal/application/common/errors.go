package common

import dc "go-ddd-book-manager/internal/domain/common"

var PaginationInvalidErr = dc.NewValidationError("pagination", "invalid")
