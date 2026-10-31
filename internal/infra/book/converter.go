package book

import (
	"go-ddd-book-manager/internal/common/option"
	domain "go-ddd-book-manager/internal/domain/book"
	du "go-ddd-book-manager/internal/domain/user"
	ic "go-ddd-book-manager/internal/infra/common"

	"github.com/samber/lo"
)

func bookModelFromDomain(b domain.Book) model {
	result := model{
		Audit:       ic.AuditFromDomain(b.Audit()),
		Title:       b.Title(),
		ISBN:        b.ISBN().Value(),
		Description: b.Description().Value(),
		OwnerID:     b.Owner().ID(),
		OwnerName:   b.Owner().Name(),
	}
	result.Authors = lo.Map(b.Authors(), func(a domain.Author, _ int) authorModel {
		return authorModel{BookID: result.ID, Name: a.Name(), OrderIndex: a.OrderIndex()}
	})
	result.Checkouts = lo.Map(b.Checkouts(), func(c domain.Checkout, index int) checkoutModel {
		return checkoutModel{
			ID:           c.ID(),
			BookID:       result.ID,
			UserID:       c.CheckedOutTo().ID(),
			UserName:     c.CheckedOutTo().Name(),
			CheckedOutAt: c.CheckedOutAt(),
			ReturnedAt:   c.ReturnedAt().Value(),
			Position:     index,
		}
	})
	return result
}

func restoreBook(m model) (domain.Book, error) {
	audit, err := m.Audit.Restore()
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
