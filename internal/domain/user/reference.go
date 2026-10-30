package user

import "uuid"

type Reference struct {
	id   uuid.UUID
	name string
}

func NewReference(id uuid.UUID, name Name) Reference {
	return Reference{id: id, name: name.value}
}
func RestoreReference(id uuid.UUID, name string) Reference {
	return Reference{id: id, name: name}
}

func (r Reference) ID() uuid.UUID { return r.id }
func (r Reference) Name() string  { return r.name }
