package common

import (
	"go-ddd-book-manager/internal/common/option"
	domain "go-ddd-book-manager/internal/domain/common"
)

func AuditFromDomain(a domain.EntityAudit) Audit {
	created := a.Created()
	result := Audit{
		ID:            a.ID(),
		CreatedByID:   created.UserID(),
		CreatedByName: created.UserName(),
		CreatedAt:     created.Timestamp(),
	}
	if updated, ok := a.Updated().Get(); ok {
		result.UpdatedByID = new(updated.UserID())
		result.UpdatedByName = new(updated.UserName())
		result.UpdatedAt = new(updated.Timestamp())
	}
	return result
}

func (a Audit) Restore() (domain.EntityAudit, error) {
	updated, err := domain.RestoreAuditRecordOption(
		option.New(a.UpdatedByID),
		option.New(a.UpdatedByName),
		option.New(a.UpdatedAt),
	)
	if err != nil {
		return domain.EntityAudit{}, err
	}
	return domain.RestoreEntityAudit(
		a.ID,
		domain.RestoreAuditRecord(a.CreatedByID, a.CreatedByName, a.CreatedAt),
		updated,
	), nil
}
