package query

import (
	"context"
	dto "go-ddd-book-manager/internal/application/user/dto"
	"uuid"
)

type QueryService interface {
	FindByID(context.Context, uuid.UUID) (*dto.UserResponse, error)
}
