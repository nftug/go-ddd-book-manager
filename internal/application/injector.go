package application

import (
	"github.com/samber/do/v2"
	bc "go-ddd-book-manager/internal/application/book/command"
	bq "go-ddd-book-manager/internal/application/book/query"
	uc "go-ddd-book-manager/internal/application/user/command"
	uq "go-ddd-book-manager/internal/application/user/query"
)

func Inject(i do.Injector) do.Injector {
	do.Provide(i, uc.NewSyncUser)
	do.Provide(i, uq.NewGetUserDetails)
	do.Provide(i, bc.NewCreateBook)
	do.Provide(i, bc.NewUpdateBook)
	do.Provide(i, bc.NewDeleteBook)
	do.Provide(i, bc.NewChangeOwner)
	do.Provide(i, bc.NewCheckoutBook)
	do.Provide(i, bc.NewReturnBook)
	do.Provide(i, bq.NewGetBookDetails)
	do.Provide(i, bq.NewGetBookList)
	do.Provide(i, bq.NewGetCheckoutHistory)
	return i
}
