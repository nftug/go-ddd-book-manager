package command

import (
	"context"
	"github.com/samber/do/v2"
	dto "go-ddd-book-manager/internal/application/book/dto"
	ac "go-ddd-book-manager/internal/application/common"
	"go-ddd-book-manager/internal/common/option"
	db "go-ddd-book-manager/internal/domain/book"
	dc "go-ddd-book-manager/internal/domain/common"
)

type UpdateBook struct {
	books db.Repository
	clock dc.Clock
	ids   dc.UUIDGenerator
}

func NewUpdateBook(i do.Injector) (*UpdateBook, error) {
	return &UpdateBook{
		books: do.MustInvoke[db.Repository](i),
		clock: do.MustInvoke[dc.Clock](i),
		ids:   do.MustInvoke[dc.UUIDGenerator](i),
	}, nil
}

func (s *UpdateBook) Execute(
	ctx context.Context,
	actor dc.Actor,
	id dto.BookIdentity,
	r dto.BookRequest,
) error {
	b, err := ac.FindRequired(func() (*db.Book, error) {
		return s.books.FindByID(ctx, id.BookID)
	})
	if err != nil {
		return err
	}
	if err = b.Update(
		dc.NewDomainContext(actor, s.ids, s.clock),
		r.Title,
		r.Authors,
		option.New(r.ISBN),
		option.New(r.Description),
	); err != nil {
		return err
	}
	return s.books.Update(ctx, *b)
}
