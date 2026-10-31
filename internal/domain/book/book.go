package book

import (
	"go-ddd-book-manager/internal/common/option"
	"go-ddd-book-manager/internal/domain/common"
	"go-ddd-book-manager/internal/domain/user"
	"uuid"
)

type Book struct {
	audit       common.EntityAudit
	title       Title
	authors     AuthorList
	isbn        option.Option[ISBN]
	description option.Option[Description]
	owner       user.Reference
	checkouts   CheckoutList
}

func Restore(
	audit common.EntityAudit,
	title string,
	authors []Author,
	isbn, description option.Option[string],
	owner user.Reference,
	checkouts []Checkout,
) Book {
	return Book{
		audit:       audit,
		title:       RestoreTitle(title),
		authors:     RestoreAuthorList(authors),
		isbn:        isbn.Map(RestoreISBN),
		description: description.Map(RestoreDescription),
		owner:       owner,
		checkouts:   RestoreCheckoutList(checkouts),
	}
}

// bookValues constructs all validated values before the aggregate is changed.
type bookValues struct {
	title       Title
	authors     AuthorList
	isbn        option.Option[ISBN]
	description option.Option[Description]
}

func newBookValues(
	title string,
	authors []string,
	isbn, description option.Option[string],
) (bookValues, error) {
	var values bookValues
	var err error
	if values.title, err = NewTitle(title); err != nil {
		return bookValues{}, err
	}
	if values.authors, err = NewAuthorList(authors); err != nil {
		return bookValues{}, err
	}
	if raw, ok := isbn.Get(); ok {
		value, err := NewISBN(raw)
		if err != nil {
			return bookValues{}, err
		}
		values.isbn = option.Some(value)
	}
	if raw, ok := description.Get(); ok {
		value, err := NewDescription(raw)
		if err != nil {
			return bookValues{}, err
		}
		values.description = option.Some(value)
	}
	return values, nil
}

func New(
	ctx common.DomainContext,
	title string,
	authors []string,
	isbn option.Option[string],
	description option.Option[string],
	owner user.Reference,
) (Book, error) {
	permission := bookPermission{actor: ctx.Actor(), ownerID: owner.ID()}
	if err := common.EnsureCanCreate(permission); err != nil {
		return Book{}, err
	}
	values, err := newBookValues(title, authors, isbn, description)
	if err != nil {
		return Book{}, err
	}
	audit, err := common.NewEntityAudit(ctx, permission)
	if err != nil {
		return Book{}, err
	}
	return Book{
		audit:       audit,
		title:       values.title,
		authors:     values.authors,
		isbn:        values.isbn,
		description: values.description,
		owner:       owner,
	}, nil
}

func (b Book) Audit() common.EntityAudit          { return b.audit }
func (b Book) ID() uuid.UUID                      { return b.audit.ID() }
func (b Book) Title() string                      { return b.title.Value() }
func (b Book) Authors() []Author                  { return b.authors.Values() }
func (b Book) ISBN() option.Option[string]        { return b.isbn.Map(ISBN.Value) }
func (b Book) Description() option.Option[string] { return b.description.Map(Description.Value) }
func (b Book) Owner() user.Reference              { return b.owner }
func (b Book) Checkouts() []Checkout              { return b.checkouts.Values() }
func (b Book) IsCheckedOut() bool                 { return b.checkouts.IsCheckedOut() }

func (b *Book) Update(
	ctx common.DomainContext,
	title string,
	authors []string,
	isbn option.Option[string],
	description option.Option[string],
) error {
	permission := bookPermission{actor: ctx.Actor(), ownerID: b.owner.ID()}
	if err := common.EnsureCanUpdate(permission); err != nil {
		return err
	}
	values, err := newBookValues(title, authors, isbn, description)
	if err != nil {
		return err
	}
	if err := b.audit.Update(ctx, permission); err != nil {
		return err
	}
	b.title, b.authors, b.isbn, b.description = values.title, values.authors, values.isbn, values.description
	return nil
}

func (b Book) ValidateDeletion(actor common.Actor) error {
	permission := bookPermission{actor: actor, ownerID: b.owner.ID()}
	return common.EnsureCanDelete(permission)
}

func (b *Book) ChangeOwner(ctx common.DomainContext, owner user.Reference) error {
	if b.owner == owner {
		return OwnerUnchangedErr
	}
	if err := b.audit.Update(ctx, ownerChangePermission{actor: ctx.Actor()}); err != nil {
		return err
	}
	b.owner = owner
	return nil
}

func (b *Book) Checkout(ctx common.DomainContext) error { return b.checkouts.checkout(ctx) }

func (b *Book) Return(ctx common.DomainContext) error { return b.checkouts.returnBook(ctx) }
