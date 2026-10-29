package common

import (
	"time"
	"uuid"
)

type UUIDGenerator interface {
	NewUUIDv7() uuid.UUID
}

type Clock interface {
	Now() time.Time
}
