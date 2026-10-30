package user

import (
	"context"

	"uuid"
)

type Repository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	Save(ctx context.Context, user User) error
}
