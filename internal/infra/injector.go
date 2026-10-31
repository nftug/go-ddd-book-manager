package infra

import (
	"github.com/samber/do/v2"
	bookinfra "go-ddd-book-manager/internal/infra/book"
	commoninfra "go-ddd-book-manager/internal/infra/common"
	userinfra "go-ddd-book-manager/internal/infra/user"
)

// NewInjector registers infrastructure services in the supplied injector.
// Shut down the injector when finished.
func NewInjector(i do.Injector, dsn string) do.Injector {
	do.ProvideValue(i, commoninfra.DatabaseConfig{DSN: dsn})
	do.Provide(i, commoninfra.NewDatabase)
	do.Provide(i, commoninfra.NewBunDB)
	do.Provide(i, bookinfra.NewRepository)
	do.Provide(i, userinfra.NewRepository)
	do.Provide(i, commoninfra.NewClock)
	do.Provide(i, commoninfra.NewUUIDGenerator)
	return i
}
