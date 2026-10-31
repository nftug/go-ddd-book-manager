package migrations

import (
	"embed"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/migrate"
)

//go:embed *.sql
var files embed.FS

func New(db *bun.DB) (*migrate.Migrator, error) {
	m := migrate.NewMigrations(migrate.WithMigrationsDirectory("internal/infra/migrations"))
	if err := m.Discover(files); err != nil {
		return nil, err
	}
	return migrate.NewMigrator(db, m), nil
}
