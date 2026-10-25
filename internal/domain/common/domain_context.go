package common

type DomainContext struct {
	actor       Actor
	idGenerator UUIDGenerator
	clock       Clock
}

func (c DomainContext) Actor() Actor               { return c.actor }
func (c DomainContext) IDGenerator() UUIDGenerator { return c.idGenerator }
func (c DomainContext) Clock() Clock               { return c.clock }

func NewDomainContext(actor Actor, idGenerator UUIDGenerator, clock Clock) DomainContext {
	return DomainContext{actor, idGenerator, clock}
}
