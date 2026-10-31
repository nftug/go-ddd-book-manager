package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"go-ddd-book-manager/internal/infra/database"
	"go-ddd-book-manager/internal/infra/migrations"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	if len(os.Args) < 2 {
		return fmt.Errorf(
			"usage: migrate <up|down|create> [name]; DATABASE_URL is required for up/down",
		)
	}
	db := database.OpenDB(os.Getenv("DATABASE_URL"))
	defer db.Close()
	m, err := migrations.New(db)
	if err != nil {
		return err
	}
	ctx := context.Background()
	switch os.Args[1] {
	case "create":
		if len(os.Args) != 3 {
			return fmt.Errorf("usage: migrate create <name>")
		}
		files, err := m.CreateSQLMigrations(ctx, os.Args[2])
		if err != nil {
			return err
		}
		for _, f := range files {
			fmt.Println(f.Path)
		}
		return nil
	case "up", "down":
		if os.Getenv("DATABASE_URL") == "" {
			return fmt.Errorf("DATABASE_URL is required")
		}
		if err = m.Init(ctx); err != nil {
			return err
		}
		if err = m.Lock(ctx); err != nil {
			return err
		}
		defer m.Unlock(ctx)
		if os.Args[1] == "up" {
			_, err = m.Migrate(ctx)
		} else {
			_, err = m.Rollback(ctx)
		}
		return err
	default:
		return fmt.Errorf("unknown command %q", os.Args[1])
	}
}
