package command

import (
	"context"
	dto "go-ddd-book-manager/internal/application/book/dto"
	ac "go-ddd-book-manager/internal/application/common"
	db "go-ddd-book-manager/internal/domain/book"
	dc "go-ddd-book-manager/internal/domain/common"
	du "go-ddd-book-manager/internal/domain/user"

	"github.com/samber/do/v2"
)

type ChangeOwner struct {
	books db.Repository
	users du.Repository
	clock dc.Clock
	ids   dc.UUIDGenerator
}

func NewChangeOwner(i do.Injector) (*ChangeOwner, error) {
	return &ChangeOwner{
		books: do.MustInvoke[db.Repository](i),
		users: do.MustInvoke[du.Repository](i),
		clock: do.MustInvoke[dc.Clock](i),
		ids:   do.MustInvoke[dc.UUIDGenerator](i),
	}, nil
}

func (s *ChangeOwner) Execute(
	ctx context.Context,
	actor dc.Actor,
	id dto.BookIdentity,
	req dto.OwnerChangeRequest,
) error {
	b, err := ac.FindRequired(func() (*db.Book, error) {
		return s.books.FindByID(ctx, id.BookID)
	})
	if err != nil {
		return err
	}

	u, err := s.users.FindByID(ctx, req.OwnerID)
	if err != nil {
		return err
	}
	if u == nil {
		return OwnerNotFoundErr
	}

	if err = b.ChangeOwner(
		dc.NewDomainContext(actor, s.ids, s.clock),
		u.Reference(),
	); err != nil {
		return err
	}
	return s.books.Update(ctx, *b)
}
