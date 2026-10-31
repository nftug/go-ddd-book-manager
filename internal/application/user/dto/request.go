package dto

import (
	dc "go-ddd-book-manager/internal/domain/common"
	"uuid"
)

type SyncUserRequest struct {
	ID    uuid.UUID    `json:"id"`
	Name  string       `json:"name"`
	Email string       `json:"email"`
	Role  dc.ActorRole `json:"role"`
}
