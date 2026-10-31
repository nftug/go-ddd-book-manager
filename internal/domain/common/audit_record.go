package common

import (
	"errors"
	"go-ddd-book-manager/internal/common/option"
	"time"
	"uuid"
)

type AuditRecord struct {
	userId    uuid.UUID
	userName  string
	timestamp time.Time
}

func (a AuditRecord) UserID() uuid.UUID    { return a.userId }
func (a AuditRecord) UserName() string     { return a.userName }
func (a AuditRecord) Timestamp() time.Time { return a.timestamp }

func RestoreAuditRecord(userId uuid.UUID, userName string, timestamp time.Time) AuditRecord {
	return AuditRecord{userId, userName, timestamp.UTC()}
}

func RestoreAuditRecordOption(
	userId option.Option[uuid.UUID],
	userName option.Option[string],
	timestamp option.Option[time.Time],
) (option.Option[AuditRecord], error) {
	if userId.IsNone() && userName.IsNone() && timestamp.IsNone() {
		return option.None[AuditRecord](), nil
	}
	if userId.IsSome() && userName.IsSome() && timestamp.IsSome() {
		userId, _ := userId.Get()
		userName, _ := userName.Get()
		timestamp, _ := timestamp.Get()
		return option.Some(AuditRecord{userId, userName, timestamp.UTC()}), nil
	}
	return option.None[AuditRecord](), errors.New("invalid state for audit record")
}

func NewAuditRecordFor(actor Actor, timestamp time.Time) AuditRecord {
	return AuditRecord{userId: actor.ID(), userName: actor.UserName(), timestamp: timestamp.UTC()}
}

func NewAuditRecord(ctx DomainContext) AuditRecord {
	return NewAuditRecordFor(ctx.Actor(), ctx.UTCNow())
}
