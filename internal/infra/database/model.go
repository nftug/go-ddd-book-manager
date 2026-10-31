package database

import (
	"time"
	"uuid"
)

type Audit struct {
	Version       int64      `bun:"version,notnull"`
	ID            uuid.UUID  `bun:"id,pk,type:uuid"`
	CreatedByID   uuid.UUID  `bun:"created_by_id,type:uuid,notnull"`
	CreatedByName string     `bun:"created_by_name,notnull"`
	CreatedAt     time.Time  `bun:"created_at,notnull"`
	UpdatedByID   *uuid.UUID `bun:"updated_by_id,type:uuid"`
	UpdatedByName *string    `bun:"updated_by_name"`
	UpdatedAt     *time.Time `bun:"updated_at"`
}
