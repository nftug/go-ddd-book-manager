package common

import (
	"uuid"
)

type ActorRole string

const (
	AdminUser  = ActorRole("Admin")
	NormalUser = ActorRole("User")
)

type Actor struct {
	id       uuid.UUID
	userName string
	role     ActorRole
}

func (a Actor) ID() uuid.UUID    { return a.id }
func (a Actor) UserName() string { return a.userName }
func (a Actor) Role() ActorRole  { return a.role }

func NewActor(id uuid.UUID, userName string, role ActorRole) Actor {
	return Actor{
		id:       id,
		userName: userName,
		role:     role,
	}
}
