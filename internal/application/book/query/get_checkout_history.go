package query

import (
	"context"
	"github.com/samber/do/v2"
	dto "go-ddd-book-manager/internal/application/book/dto"
	ac "go-ddd-book-manager/internal/application/common"
	dc "go-ddd-book-manager/internal/domain/common"
)

type GetCheckoutHistory struct{ q QueryService }

func NewGetCheckoutHistory(i do.Injector) (*GetCheckoutHistory, error) {
	return &GetCheckoutHistory{q: do.MustInvoke[QueryService](i)}, nil
}

func (s *GetCheckoutHistory) Execute(
	ctx context.Context,
	actor dc.Actor,
	id dto.BookIdentity,
	p ac.Pagination,
) (dto.CheckoutPage, error) {
	if !actor.IsAdmin() {
		return dto.CheckoutPage{}, dc.ErrForbidden
	}
	p, err := p.Normalize()
	if err != nil {
		return dto.CheckoutPage{}, err
	}
	page, err := ac.FindRequired(func() (*dto.CheckoutPage, error) {
		return s.q.Checkouts(ctx, id.BookID, p)
	})
	if err != nil {
		return dto.CheckoutPage{}, err
	}
	return *page, nil
}
