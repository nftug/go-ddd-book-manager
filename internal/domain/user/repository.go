package user

import (
	"context"
	"uuid"
)

type Repository interface {
	FindByID(context.Context, uuid.UUID) (*User, error)
	Create(context.Context, User) error
	Update(context.Context, User) error
	Delete(context.Context, User) error
}
