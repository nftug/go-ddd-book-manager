package auth

import (
	"context"
	dc "go-ddd-book-manager/internal/domain/common"
	"uuid"
)

type Identity struct {
	ID          uuid.UUID
	Name, Email string
	Role        dc.ActorRole
}

type TokenVerifier interface {
	Verify(context.Context, string) (Identity, error)
}
