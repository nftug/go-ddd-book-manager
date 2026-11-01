package common

import (
	"time"
	"uuid"

	domain "go-ddd-book-manager/internal/domain/common"
)

type UserReference struct {
	ID   uuid.UUID `json:"id" format:"uuid"`
	Name string    `json:"name"`
} // @name UserReference

type AuditRecord struct {
	User UserReference `json:"user"`
	At   time.Time     `json:"at" format:"date-time"`
} // @name AuditRecord

type Audit struct {
	Created AuditRecord  `json:"created"`
	Updated *AuditRecord `json:"updated"`
} // @name Audit

func AuditFromDomain(a domain.EntityAudit) Audit {
	created := a.Created()
	result := Audit{
		Created: AuditRecord{
			User: UserReference{ID: created.UserID(), Name: created.UserName()},
			At:   created.Timestamp(),
		},
	}
	if updated, ok := a.Updated().Get(); ok {
		result.Updated = &AuditRecord{
			User: UserReference{ID: updated.UserID(), Name: updated.UserName()},
			At:   updated.Timestamp(),
		}
	}
	return result
}
