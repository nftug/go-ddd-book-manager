package webapi

import (
	"github.com/samber/do/v2"
	"go-ddd-book-manager/internal/webapi/auth"
	"go-ddd-book-manager/internal/webapi/book"
	"go-ddd-book-manager/internal/webapi/user"
)

func Inject(i do.Injector, config auth.Config) do.Injector {
	do.ProvideValue(i, config)
	do.Provide(i, auth.NewVerifier)
	do.Provide(i, book.NewHandler)
	do.Provide(i, user.NewHandler)
	do.Provide(i, New)

	return i
}
