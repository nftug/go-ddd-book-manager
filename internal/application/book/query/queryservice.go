package query

import (
	"context"
	dto "go-ddd-book-manager/internal/application/book/dto"
	ac "go-ddd-book-manager/internal/application/common"
	dc "go-ddd-book-manager/internal/domain/common"
	"uuid"
)

type QueryService interface {
	FindByID(context.Context, dc.Actor, uuid.UUID) (*dto.BookResponse, error)
	// List and Checkouts require normalized pagination.
	List(context.Context, dc.Actor, dto.BookQuery) (dto.BookPage, error)
	Checkouts(context.Context, uuid.UUID, ac.Pagination) (*dto.CheckoutPage, error)
}
