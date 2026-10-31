package book

import (
	"context"
	"uuid"
)

type Repository interface {
	FindByID(context.Context, uuid.UUID) (*Book, error)
	Create(context.Context, Book) error
	Update(context.Context, Book) error
	Delete(context.Context, Book) error
}
