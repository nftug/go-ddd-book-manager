package common

import (
	"uuid"
)

type ActorRole string

const (
	AdminRole = ActorRole("Admin")
	UserRole  = ActorRole("User")
)

type Actor struct {
	id       uuid.UUID
	userName string
	role     ActorRole
	system   bool
}

func (a Actor) ID() uuid.UUID    { return a.id }
func (a Actor) UserName() string { return a.userName }
func (a Actor) Role() ActorRole  { return a.role }
func (a Actor) IsSystem() bool   { return a.system }
func (a Actor) IsAdmin() bool    { return !a.system && a.role == AdminRole }

func NewActor(id uuid.UUID, userName string, role ActorRole) Actor {
	return Actor{id, userName, role, false}
}

// NewSystemActor creates the actor used for internal automated operations.
func NewSystemActor() Actor {
	return Actor{id: uuid.Nil(), userName: "System", system: true}
}
