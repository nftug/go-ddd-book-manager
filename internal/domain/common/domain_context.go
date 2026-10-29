package common

import (
	"time"
	"uuid"
)

type DomainContext struct {
	actor       Actor
	idGenerator UUIDGenerator
	clock       Clock
}

func (c DomainContext) Actor() Actor         { return c.actor }
func (c DomainContext) NewUUIDv7() uuid.UUID { return c.idGenerator.NewUUIDv7() }
func (c DomainContext) UTCNow() time.Time    { return c.clock.Now().UTC() }

func NewDomainContext(actor Actor, idGenerator UUIDGenerator, clock Clock) DomainContext {
	return DomainContext{actor, idGenerator, clock}
}
