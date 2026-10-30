package user

import (
	"go-ddd-book-manager/internal/domain/common"
	"uuid"
)

// userPermission describes which user profile an actor may synchronize.
type userPermission struct {
	actor  common.Actor
	userID uuid.UUID
}

func (p userPermission) CanCreate() bool { return false }
func (p userPermission) CanUpdate() bool {
	return p.actor.IsSystem() || p.actor.ID() == p.userID
}
func (p userPermission) CanDelete() bool { return false }

type userCretePermission struct{}

func (userCretePermission) CanCreate() bool { return true }
func (userCretePermission) CanUpdate() bool { return false }
func (userCretePermission) CanDelete() bool { return false }
