package command

import (
	"context"
	dto "go-ddd-book-manager/internal/application/book/dto"
	"go-ddd-book-manager/internal/application/common"
	"go-ddd-book-manager/internal/common/option"
	db "go-ddd-book-manager/internal/domain/book"
	dc "go-ddd-book-manager/internal/domain/common"
	du "go-ddd-book-manager/internal/domain/user"

	"github.com/samber/do/v2"
)

type CreateBook struct {
	books db.Repository
	clock dc.Clock
	ids   dc.UUIDGenerator
}

func NewCreateBook(i do.Injector) (*CreateBook, error) {
	return &CreateBook{
		books: do.MustInvoke[db.Repository](i),
		clock: do.MustInvoke[dc.Clock](i),
		ids:   do.MustInvoke[dc.UUIDGenerator](i),
	}, nil
}

func (s *CreateBook) Execute(
	ctx context.Context,
	actor dc.Actor,
	r dto.BookRequest,
) (common.CreatedResponse, error) {
	b, err := db.New(
		dc.NewDomainContext(actor, s.ids, s.clock),
		r.Title,
		r.Authors,
		option.New(r.ISBN),
		option.New(r.Description),
		du.RestoreReference(actor.ID(), actor.UserName()),
	)
	if err != nil {
		return common.CreatedResponse{}, err
	}
	if err = s.books.Create(ctx, b); err != nil {
		return common.CreatedResponse{}, err
	}
	return common.CreatedResponse{ID: b.ID()}, nil
}
