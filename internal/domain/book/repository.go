package book

import (
	"context"

	"uuid"
)

type Repository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*Book, error)
	Save(ctx context.Context, book Book) error
	Delete(ctx context.Context, id uuid.UUID) error
}
