package common

import (
	"time"
	"uuid"
)

type UUIDGenerator interface {
	New() uuid.UUID
}

type Clock interface {
	Now() time.Time
}
