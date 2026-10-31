package command

import (
	"context"
	"github.com/samber/do/v2"
	dto "go-ddd-book-manager/internal/application/book/dto"
	ac "go-ddd-book-manager/internal/application/common"
	db "go-ddd-book-manager/internal/domain/book"
	dc "go-ddd-book-manager/internal/domain/common"
)

type DeleteBook struct{ books db.Repository }

func NewDeleteBook(i do.Injector) (*DeleteBook, error) {
	return &DeleteBook{books: do.MustInvoke[db.Repository](i)}, nil
}

func (s *DeleteBook) Execute(ctx context.Context, actor dc.Actor, id dto.BookIdentity) error {
	b, err := ac.FindRequired(func() (*db.Book, error) {
		return s.books.FindByID(ctx, id.BookID)
	})
	if err != nil {
		return err
	}
	if err = b.ValidateDeletion(actor); err != nil {
		return err
	}
	return s.books.Delete(ctx, *b)
}
