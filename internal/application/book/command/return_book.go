package command

import (
	"context"
	"github.com/samber/do/v2"
	dto "go-ddd-book-manager/internal/application/book/dto"
	ac "go-ddd-book-manager/internal/application/common"
	db "go-ddd-book-manager/internal/domain/book"
	dc "go-ddd-book-manager/internal/domain/common"
)

type ReturnBook struct {
	books db.Repository
	clock dc.Clock
	ids   dc.UUIDGenerator
}

func NewReturnBook(i do.Injector) (*ReturnBook, error) {
	return &ReturnBook{
		books: do.MustInvoke[db.Repository](i),
		clock: do.MustInvoke[dc.Clock](i),
		ids:   do.MustInvoke[dc.UUIDGenerator](i),
	}, nil
}

func (s *ReturnBook) Execute(ctx context.Context, actor dc.Actor, id dto.BookIdentity) error {
	b, err := ac.FindRequired(func() (*db.Book, error) {
		return s.books.FindByID(ctx, id.BookID)
	})
	if err != nil {
		return err
	}
	if err = b.Return(dc.NewDomainContext(actor, s.ids, s.clock)); err != nil {
		return err
	}
	return s.books.Update(ctx, *b)
}
