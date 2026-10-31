package common

//go:generate go tool go-enum --noprefix

import (
	"uuid"
)

// ENUM(AdminRole="Admin", UserRole="User", SystemRole="System")
type ActorRole string

type Actor struct {
	id       uuid.UUID
	userName string
	role     ActorRole
}

func (a Actor) ID() uuid.UUID    { return a.id }
func (a Actor) UserName() string { return a.userName }
func (a Actor) Role() ActorRole  { return a.role }
func (a Actor) IsSystem() bool   { return a.role == SystemRole }
func (a Actor) IsAdmin() bool    { return a.role == AdminRole }

func NewActor(id uuid.UUID, userName string, role ActorRole) Actor {
	return Actor{id, userName, role}
}

// NewSystemActor creates the actor used for internal automated operations.
func NewSystemActor() Actor {
	return Actor{id: uuid.Nil(), userName: "System", role: SystemRole}
}
