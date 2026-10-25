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

func (a AuditRecord) UserId() uuid.UUID    { return a.userId }
func (a AuditRecord) UserName() string     { return a.userName }
func (a AuditRecord) Timestamp() time.Time { return a.timestamp }

func RestoreAuditRecord(userId uuid.UUID, userName string, timestamp time.Time) AuditRecord {
	return AuditRecord{userId, userName, timestamp}
}

func RestoreAuditRecordOption(
	userId *uuid.UUID,
	userName *string,
	timestamp *time.Time,
) (option.Option[AuditRecord], error) {
	if userId == nil && userName == nil && timestamp == nil {
		return option.None[AuditRecord](), nil
	}
	if userId != nil && userName != nil && timestamp != nil {
		return option.Some(AuditRecord{*userId, *userName, *timestamp}), nil
	}
	return option.None[AuditRecord](), errors.New("invalid state for audit record")
}

func NewAuditRecord(ctx DomainContext) AuditRecord {
	return AuditRecord{
		userId:    ctx.Actor().ID(),
		userName:  ctx.Actor().UserName(),
		timestamp: ctx.Clock().Now(),
	}
}
