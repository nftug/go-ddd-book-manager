package book

import (
	"time"
	"uuid"

	"go-ddd-book-manager/internal/infra/database"

	"github.com/uptrace/bun"
)

type model struct {
	bun.BaseModel  `bun:"table:books"`
	database.Audit `bun:",embed"`
	Title          string          `bun:"title,notnull"`
	ISBN           *string         `bun:"isbn"`
	Description    *string         `bun:"description"`
	OwnerID        uuid.UUID       `bun:"owner_id,type:uuid,notnull"`
	OwnerName      string          `bun:"owner_name,notnull"`
	Authors        []authorModel   `bun:"rel:has-many,join:id=book_id"`
	Checkouts      []checkoutModel `bun:"rel:has-many,join:id=book_id"`
}

type authorModel struct {
	bun.BaseModel `bun:"table:book_authors"`
	BookID        uuid.UUID `bun:"book_id,type:uuid"`
	Name          string    `bun:"name,notnull"`
	OrderIndex    int       `bun:"order_index,notnull"`
}

type checkoutModel struct {
	bun.BaseModel `bun:"table:book_checkouts"`
	ID            uuid.UUID  `bun:"id,type:uuid"`
	BookID        uuid.UUID  `bun:"book_id,type:uuid"`
	UserID        uuid.UUID  `bun:"user_id,type:uuid"`
	UserName      string     `bun:"user_name,notnull"`
	CheckedOutAt  time.Time  `bun:"checked_out_at,notnull"`
	ReturnedAt    *time.Time `bun:"returned_at"`
	Position      int        `bun:"position,notnull"`
}

func relations(q *bun.SelectQuery) *bun.SelectQuery {
	return q.
		Relation("Authors", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order("order_index ASC")
		}).
		Relation("Checkouts", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order("position ASC")
		})
}
