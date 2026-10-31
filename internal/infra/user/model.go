package user

import (
	dc "go-ddd-book-manager/internal/domain/common"
	ic "go-ddd-book-manager/internal/infra/common"

	"github.com/uptrace/bun"
)

type model struct {
	bun.BaseModel `bun:"table:users"`
	ic.Audit      `bun:",embed"`
	Name          string       `bun:"name,notnull"`
	Email         string       `bun:"email,notnull"`
	Role          dc.ActorRole `bun:"role,notnull"`
}
