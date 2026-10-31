package database

import (
	"database/sql"

	"github.com/samber/do/v2"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

type DatabaseConfig struct{ DSN string }
type Database struct{ db *bun.DB }

func (d *Database) Shutdown() error { return d.db.Close() }

func NewDatabase(i do.Injector) (*Database, error) {
	config := do.MustInvoke[DatabaseConfig](i)
	return &Database{db: OpenDB(config.DSN)}, nil
}

func NewBunDB(i do.Injector) (*bun.DB, error) {
	database := do.MustInvoke[*Database](i)
	return database.db, nil
}

func OpenDB(dsn string) *bun.DB {
	return bun.NewDB(sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn))), pgdialect.New())
}
