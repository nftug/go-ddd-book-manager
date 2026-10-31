package user

import (
	dc "go-ddd-book-manager/internal/domain/common"
	"go-ddd-book-manager/internal/infra/database"

	"github.com/uptrace/bun"
)

type model struct {
	bun.BaseModel  `bun:"table:users"`
	database.Audit `bun:",embed"`
	Name           string       `bun:"name,notnull"`
	Email          string       `bun:"email,notnull"`
	Role           dc.ActorRole `bun:"role,notnull"`
}
