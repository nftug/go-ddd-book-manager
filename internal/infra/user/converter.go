package user

import (
	dto "go-ddd-book-manager/internal/application/user/dto"
	domain "go-ddd-book-manager/internal/domain/user"
	"go-ddd-book-manager/internal/infra/database"
)

func (m *model) fromDomain(u domain.User) {
	var audit database.Audit
	audit.FromDomain(u.Audit())
	*m = model{
		Audit: audit,
		Name:  u.Name(),
		Email: u.Email(),
		Role:  u.Role(),
	}
}

func (m model) toDomain() (domain.User, error) {
	audit, err := m.Audit.RestoreAudit()
	if err != nil {
		return domain.User{}, err
	}
	return domain.Restore(audit, m.Name, m.Email, m.Role)
}

func (m model) toResponse() dto.UserResponse {
	return dto.UserResponse{
		ID:    m.ID,
		Name:  m.Name,
		Email: m.Email,
		Role:  m.Role,
		Audit: m.Audit.ToResponse(),
	}
}
