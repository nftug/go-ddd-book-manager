package book

import (
	"time"

	"go-ddd-book-manager/internal/common/option"
	"go-ddd-book-manager/internal/domain/common"
	"go-ddd-book-manager/internal/domain/user"
	"uuid"

	"github.com/samber/lo"
)

type Checkout struct {
	id           uuid.UUID
	checkedOutTo user.Reference
	checkedOutAt time.Time
	returnedAt   option.Option[time.Time]
}

func RestoreCheckout(
	id uuid.UUID,
	checkedOutTo user.Reference,
	checkedOutAt time.Time,
	returnedAt option.Option[time.Time],
) Checkout {
	return Checkout{
		id:           id,
		checkedOutTo: checkedOutTo,
		checkedOutAt: checkedOutAt.UTC(),
		returnedAt:   returnedAt.Map(func(value time.Time) time.Time { return value.UTC() }),
	}
}

func (c Checkout) ID() uuid.UUID                        { return c.id }
func (c Checkout) CheckedOutTo() user.Reference         { return c.checkedOutTo }
func (c Checkout) CheckedOutAt() time.Time              { return c.checkedOutAt }
func (c Checkout) ReturnedAt() option.Option[time.Time] { return c.returnedAt }
func (c Checkout) IsActive() bool                       { return c.returnedAt.IsNone() }

type CheckoutList struct{ values []Checkout }

func RestoreCheckoutList(values []Checkout) CheckoutList {
	return CheckoutList{lo.Clone(values)}
}

func (l CheckoutList) Values() []Checkout { return lo.Clone(l.values) }
func (l CheckoutList) IsCheckedOut() bool {
	latest, ok := l.latest()
	return ok && latest.IsActive()
}

func (l CheckoutList) latest() (Checkout, bool) {
	if len(l.values) == 0 {
		return Checkout{}, false
	}
	latest := lo.MaxBy(l.values, func(a, b Checkout) bool {
		return a.checkedOutAt.After(b.checkedOutAt) || a.checkedOutAt.Equal(b.checkedOutAt)
	})
	return latest, true
}

func (l *CheckoutList) checkout(ctx common.DomainContext) error {
	if l.IsCheckedOut() {
		return common.NewValidationError("book.checkout.already_active")
	}
	actor := ctx.Actor()
	if actor.IsSystem() {
		return common.ErrForbidden
	}

	to := user.RestoreReference(actor.ID(), actor.UserName())
	l.values = append(
		l.values,
		Checkout{id: ctx.NewUUIDv7(), checkedOutTo: to, checkedOutAt: ctx.UTCNow()},
	)
	return nil
}

func (l *CheckoutList) returnBook(ctx common.DomainContext) error {
	latest, ok := l.latest()
	if !ok {
		return common.NewValidationError("book.checkout.not_active")
	}
	if !latest.IsActive() {
		return common.NewValidationError("book.checkout.already_returned")
	}
	actor := ctx.Actor()
	if actor.IsSystem() || (latest.checkedOutTo.ID() != actor.ID() && !actor.IsAdmin()) {
		return common.ErrForbidden
	}

	_, index, found := lo.FindIndexOf(l.values, func(checkout Checkout) bool {
		return checkout.id == latest.id
	})
	if found {
		l.values[index].returnedAt = option.Some(ctx.UTCNow())
		return nil
	}
	return common.NewValidationError("book.checkout.not_active")
}
