package common

import (
	"database/sql"
	"time"
	"uuid"

	"github.com/samber/do/v2"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	domain "go-ddd-book-manager/internal/domain/common"
)

type Clock struct{}

func (Clock) Now() time.Time { return time.Now() }

func NewClock(do.Injector) (domain.Clock, error) { return Clock{}, nil }

type UUIDGenerator struct{}

func (UUIDGenerator) NewUUIDv7() uuid.UUID { return uuid.NewV7() }

func NewUUIDGenerator(do.Injector) (domain.UUIDGenerator, error) { return UUIDGenerator{}, nil }

func OpenDB(dsn string) *bun.DB {
	return bun.NewDB(sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn))), pgdialect.New())
}
