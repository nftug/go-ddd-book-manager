package database

import (
	ac "go-ddd-book-manager/internal/application/common"
	"go-ddd-book-manager/internal/common/option"
	domain "go-ddd-book-manager/internal/domain/common"
)

func (a Audit) ToResponse() ac.Audit {
	response := ac.Audit{
		Created: ac.AuditRecord{
			User: ac.UserReference{ID: a.CreatedByID, Name: a.CreatedByName},
			At:   a.CreatedAt.UTC(),
		},
	}
	if a.UpdatedByID != nil {
		response.Updated = &ac.AuditRecord{
			User: ac.UserReference{ID: *a.UpdatedByID, Name: *a.UpdatedByName},
			At:   a.UpdatedAt.UTC(),
		}
	}
	return response
}

func (a *Audit) FromDomain(entityAudit domain.EntityAudit) {
	created := entityAudit.Created()
	*a = Audit{
		Version:       entityAudit.Version(),
		ID:            entityAudit.ID(),
		CreatedByID:   created.UserID(),
		CreatedByName: created.UserName(),
		CreatedAt:     created.Timestamp(),
	}
	if updated, ok := entityAudit.Updated().Get(); ok {
		a.UpdatedByID = new(updated.UserID())
		a.UpdatedByName = new(updated.UserName())
		a.UpdatedAt = new(updated.Timestamp())
	}
}

func (a Audit) RestoreAudit() (domain.EntityAudit, error) {
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
		a.Version,
		domain.RestoreAuditRecord(a.CreatedByID, a.CreatedByName, a.CreatedAt),
		updated,
	), nil
}
