package command

import (
	"context"
	"github.com/samber/do/v2"
	dto "go-ddd-book-manager/internal/application/book/dto"
	ac "go-ddd-book-manager/internal/application/common"
	db "go-ddd-book-manager/internal/domain/book"
	dc "go-ddd-book-manager/internal/domain/common"
)

type CheckoutBook interface {
	Execute(ctx context.Context, actor dc.Actor, id dto.BookIdentity) error
}

type checkoutBook struct {
	books db.Repository
	clock dc.Clock
	ids   dc.UUIDGenerator
}

func NewCheckoutBook(i do.Injector) (CheckoutBook, error) {
	return &checkoutBook{
		books: do.MustInvoke[db.Repository](i),
		clock: do.MustInvoke[dc.Clock](i),
		ids:   do.MustInvoke[dc.UUIDGenerator](i),
	}, nil
}

func (s *checkoutBook) Execute(ctx context.Context, actor dc.Actor, id dto.BookIdentity) error {
	b, err := ac.FindRequired(func() (*db.Book, error) {
		return s.books.FindByID(ctx, id.BookID)
	})
	if err != nil {
		return err
	}
	if err = b.Checkout(dc.NewDomainContext(actor, s.ids, s.clock)); err != nil {
		return err
	}
	return s.books.Update(ctx, *b)
}
