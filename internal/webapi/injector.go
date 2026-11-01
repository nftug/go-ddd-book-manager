package webapi

import (
	"github.com/samber/do/v2"
	"go-ddd-book-manager/internal/webapi/auth"
	"go-ddd-book-manager/internal/webapi/book"
	"go-ddd-book-manager/internal/webapi/user"
)

type Config struct {
	AllowedOrigins []string
}

func Inject(i do.Injector, config auth.Config, webConfig Config) do.Injector {
	do.ProvideValue(i, config)
	do.ProvideValue(i, webConfig)
	do.Provide(i, auth.NewVerifier)
	do.Provide(i, book.NewHandler)
	do.Provide(i, user.NewHandler)
	do.Provide(i, New)

	return i
}
