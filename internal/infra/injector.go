package infra

import (
	bookinfra "go-ddd-book-manager/internal/infra/book"
	commoninfra "go-ddd-book-manager/internal/infra/common"
	databaseinfra "go-ddd-book-manager/internal/infra/database"
	userinfra "go-ddd-book-manager/internal/infra/user"

	"github.com/samber/do/v2"
)

func Inject(i do.Injector) do.Injector {
	do.Provide(i, databaseinfra.NewDatabase)
	do.Provide(i, databaseinfra.NewBunDB)
	do.Provide(i, bookinfra.NewRepository)
	do.Provide(i, bookinfra.NewQueryService)
	do.Provide(i, userinfra.NewRepository)
	do.Provide(i, userinfra.NewQueryService)
	do.Provide(i, commoninfra.NewClock)
	do.Provide(i, commoninfra.NewUUIDGenerator)
	return i
}
