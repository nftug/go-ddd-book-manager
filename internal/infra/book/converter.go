package book

import (
	dto "go-ddd-book-manager/internal/application/book/dto"
	ac "go-ddd-book-manager/internal/application/common"
	"go-ddd-book-manager/internal/common/option"
	domain "go-ddd-book-manager/internal/domain/book"
	dc "go-ddd-book-manager/internal/domain/common"
	du "go-ddd-book-manager/internal/domain/user"
	"go-ddd-book-manager/internal/infra/database"

	"github.com/samber/lo"
)

// From/To Domain

func (m *model) fromDomain(b domain.Book) {
	var audit database.Audit
	audit.FromDomain(b.Audit())
	*m = model{
		Audit:       audit,
		Title:       b.Title(),
		ISBN:        b.ISBN().Value(),
		Description: b.Description().Value(),
		OwnerID:     b.Owner().ID(),
		OwnerName:   b.Owner().Name(),
	}
	m.Authors = lo.Map(b.Authors(), func(a domain.Author, _ int) authorModel {
		return authorModel{BookID: m.ID, Name: a.Name(), OrderIndex: a.OrderIndex()}
	})
	m.Checkouts = lo.Map(b.Checkouts(), func(c domain.Checkout, index int) checkoutModel {
		return checkoutModel{
			ID:           c.ID(),
			BookID:       m.ID,
			UserID:       c.CheckedOutTo().ID(),
			UserName:     c.CheckedOutTo().Name(),
			CheckedOutAt: c.CheckedOutAt(),
			ReturnedAt:   c.ReturnedAt().Value(),
			Position:     index,
		}
	})
}

func (m model) toDomain() (domain.Book, error) {
	audit, err := m.Audit.RestoreAudit()
	if err != nil {
		return domain.Book{}, err
	}
	authors := lo.Map(m.Authors, func(a authorModel, _ int) domain.Author {
		return domain.RestoreAuthor(a.Name, a.OrderIndex)
	})
	checkouts := lo.Map(m.Checkouts, func(c checkoutModel, _ int) domain.Checkout {
		return domain.RestoreCheckout(
			c.ID,
			du.RestoreReference(c.UserID, c.UserName),
			c.CheckedOutAt,
			option.New(c.ReturnedAt),
		)
	})
	return domain.Restore(
		audit,
		m.Title,
		authors,
		option.New(m.ISBN),
		option.New(m.Description),
		du.RestoreReference(m.OwnerID, m.OwnerName),
		checkouts,
	), nil
}

// To Response

func (m model) toResponse(actor dc.Actor) dto.BookResponse {
	authors := lo.Map(m.Authors, func(author authorModel, _ int) string {
		return author.Name
	})
	latestCheckout := lo.MaxBy(m.Checkouts, func(a, b checkoutModel) bool {
		return lo.Ternary(
			!a.CheckedOutAt.Equal(b.CheckedOutAt),
			a.CheckedOutAt.After(b.CheckedOutAt),
			a.Position > b.Position,
		)
	})
	checkedOut := len(m.Checkouts) > 0 && latestCheckout.ReturnedAt == nil
	return dto.BookResponse{
		ID:          m.ID,
		Title:       m.Title,
		Authors:     authors,
		ISBN:        m.ISBN,
		Description: m.Description,
		Owner:       ac.UserReference{ID: m.OwnerID, Name: m.OwnerName},
		CheckedOut:  checkedOut,
		Audit:       m.Audit.ToResponse(),
		Permission:  ac.PermissionFromDomain(domain.PermissionsFor(actor, m.OwnerID)),
	}
}

func (model checkoutModel) toResponse() dto.CheckoutResponse {
	return dto.CheckoutResponse{
		ID:           model.ID,
		CheckedOutTo: ac.UserReference{ID: model.UserID, Name: model.UserName},
		CheckedOutAt: model.CheckedOutAt,
		ReturnedAt:   model.ReturnedAt,
	}
}
