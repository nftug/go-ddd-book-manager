package dto

import (
	ac "go-ddd-book-manager/internal/application/common"
	dc "go-ddd-book-manager/internal/domain/common"
	du "go-ddd-book-manager/internal/domain/user"
	"uuid"
)

type UserResponse struct {
	ID    uuid.UUID    `json:"id"`
	Name  string       `json:"name"`
	Email string       `json:"email"`
	Role  dc.ActorRole `json:"role"`
	Audit ac.Audit     `json:"audit"`
}

func FromDomain(u du.User) UserResponse {
	return UserResponse{u.ID(), u.Name(), u.Email(), u.Role(), ac.AuditFromDomain(u.Audit())}
}
