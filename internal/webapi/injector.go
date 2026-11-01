package webapi

import (
	"go-ddd-book-manager/internal/webapi/auth"
	"go-ddd-book-manager/internal/webapi/book"
	"go-ddd-book-manager/internal/webapi/user"

	"github.com/samber/do/v2"
)

func Inject(i do.Injector) do.Injector {
	do.Provide(i, auth.NewVerifier)
	do.Provide(i, book.NewHandler)
	do.Provide(i, user.NewHandler)
	do.Provide(i, New)
	return i
}
