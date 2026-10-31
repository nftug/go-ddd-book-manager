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
	authors []OrderedAuthor,
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

func New(
	ctx common.DomainContext,
	title Title,
	authors AuthorList,
	isbn option.Option[ISBN],
	description option.Option[Description],
	owner user.Reference,
) (Book, error) {
	permission := bookPermission{actor: ctx.Actor(), ownerID: owner.ID()}
	audit, err := common.NewEntityAudit(ctx, permission)
	if err != nil {
		return Book{}, err
	}
	return Book{
		audit:       audit,
		title:       title,
		authors:     authors,
		isbn:        isbn,
		description: description,
		owner:       owner,
	}, nil
}

func (b Book) Audit() common.EntityAudit          { return b.audit }
func (b Book) ID() uuid.UUID                      { return b.audit.ID() }
func (b Book) Title() string                      { return b.title.Value() }
func (b Book) Authors() []OrderedAuthor           { return b.authors.Values() }
func (b Book) ISBN() option.Option[string]        { return b.isbn.Map(ISBN.Value) }
func (b Book) Description() option.Option[string] { return b.description.Map(Description.Value) }
func (b Book) Owner() user.Reference              { return b.owner }
func (b Book) Checkouts() []Checkout              { return b.checkouts.Values() }
func (b Book) IsCheckedOut() bool                 { return b.checkouts.IsCheckedOut() }

func (b *Book) Update(
	ctx common.DomainContext,
	title Title,
	authors AuthorList,
	isbn option.Option[ISBN],
	description option.Option[Description],
) error {
	permission := bookPermission{actor: ctx.Actor(), ownerID: b.owner.ID()}
	if err := b.audit.Update(ctx, permission); err != nil {
		return err
	}
	b.title, b.authors, b.isbn, b.description = title, authors, isbn, description
	return nil
}

func (b Book) ValidateDeletion(actor common.Actor) error {
	if !canManageBook(actor, b.owner.ID()) {
		return common.ErrForbidden
	}
	return nil
}

func (b *Book) ChangeOwner(ctx common.DomainContext, owner user.Reference) error {
	if !ctx.Actor().IsAdmin() {
		return common.ErrForbidden
	}
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
