package common

import (
	"go-ddd-book-manager/internal/common/option"
	"uuid"
)

type EntityAudit struct {
	id      uuid.UUID
	created AuditRecord
	updated option.Option[AuditRecord]
}

func (e EntityAudit) ID() uuid.UUID                       { return e.id }
func (e EntityAudit) Created() AuditRecord                { return e.created }
func (e EntityAudit) Updated() option.Option[AuditRecord] { return e.updated }

func RestoreEntityAudit(
	id uuid.UUID,
	created AuditRecord,
	updated option.Option[AuditRecord],
) EntityAudit {
	return EntityAudit{id, created, updated}
}

func NewEntityAudit(ctx DomainContext, permission PermissionPolicy) (EntityAudit, error) {
	if err := EnsureCanCreate(permission); err != nil {
		return EntityAudit{}, err
	}
	return newEntityAudit(ctx, ctx.NewUUIDv7()), nil
}

// NewEntityAuditWithID creates audit data for an entity whose ID comes from an external identity provider.
func NewEntityAuditWithID(
	ctx DomainContext,
	permission PermissionPolicy,
	id uuid.UUID,
) (EntityAudit, error) {
	if err := EnsureCanCreate(permission); err != nil {
		return EntityAudit{}, err
	}
	return newEntityAudit(ctx, id), nil
}

func newEntityAudit(ctx DomainContext, id uuid.UUID) EntityAudit {
	return EntityAudit{
		id:      id,
		created: NewAuditRecord(ctx),
		updated: option.None[AuditRecord](),
	}
}

func (e *EntityAudit) Update(ctx DomainContext, permission PermissionPolicy) error {
	if err := EnsureCanUpdate(permission); err != nil {
		return err
	}
	e.updated = option.Some(NewAuditRecord(ctx))
	return nil
}
