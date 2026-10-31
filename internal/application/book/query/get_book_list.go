package query

import (
	"context"
	"github.com/samber/do/v2"
	dto "go-ddd-book-manager/internal/application/book/dto"
	dc "go-ddd-book-manager/internal/domain/common"
)

type GetBookList interface {
	Execute(
		ctx context.Context,
		actor dc.Actor,
		q dto.BookQuery,
	) (dto.BookPage, error)
}

type getBookList struct{ q QueryService }

func NewGetBookList(i do.Injector) (GetBookList, error) {
	return &getBookList{q: do.MustInvoke[QueryService](i)}, nil
}

func (s *getBookList) Execute(
	ctx context.Context,
	actor dc.Actor,
	q dto.BookQuery,
) (dto.BookPage, error) {
	p, err := q.Pagination.Normalize()
	if err != nil {
		return dto.BookPage{}, err
	}
	q.Pagination = p
	return s.q.List(ctx, actor, q)
}
