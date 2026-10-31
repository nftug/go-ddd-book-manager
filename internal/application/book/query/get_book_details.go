package query

import (
	"context"
	"github.com/samber/do/v2"
	dto "go-ddd-book-manager/internal/application/book/dto"
	ac "go-ddd-book-manager/internal/application/common"
	dc "go-ddd-book-manager/internal/domain/common"
)

type GetBookDetails struct{ q QueryService }

func NewGetBookDetails(i do.Injector) (*GetBookDetails, error) {
	return &GetBookDetails{q: do.MustInvoke[QueryService](i)}, nil
}

func (s *GetBookDetails) Execute(
	ctx context.Context,
	actor dc.Actor,
	id dto.BookIdentity,
) (dto.BookResponse, error) {
	b, err := ac.FindRequired(func() (*dto.BookResponse, error) {
		return s.q.FindByID(ctx, actor, id.BookID)
	})
	if err != nil {
		return dto.BookResponse{}, err
	}
	return *b, nil
}
