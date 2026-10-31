package user

import "go-ddd-book-manager/internal/domain/common"

var (
	NameEmptyErr    = common.NewValidationError("user.name", "empty")
	NameTooLongErr  = common.NewValidationError("user.name", "too_long")
	EmailInvalidErr = common.NewValidationError("user.email", "invalid")
	RoleInvalidErr  = common.NewValidationError("user.role", "invalid")
)
