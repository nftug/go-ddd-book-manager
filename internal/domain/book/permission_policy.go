package book

import (
	"go-ddd-book-manager/internal/domain/common"
	"uuid"
)

type bookPermission struct {
	actor   common.Actor
	ownerID uuid.UUID
}

func (p bookPermission) CanCreate() bool { return canManageBook(p.actor, p.ownerID) }
func (p bookPermission) CanUpdate() bool { return canManageBook(p.actor, p.ownerID) }
func (p bookPermission) CanDelete() bool { return canManageBook(p.actor, p.ownerID) }

type ownerChangePermission struct{ actor common.Actor }

func (p ownerChangePermission) CanCreate() bool { return false }
func (p ownerChangePermission) CanUpdate() bool { return p.actor.IsAdmin() }
func (p ownerChangePermission) CanDelete() bool { return false }

func canManageBook(actor common.Actor, ownerID uuid.UUID) bool {
	return !actor.IsSystem() && (actor.IsAdmin() || actor.ID() == ownerID)
}
