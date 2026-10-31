package user

import (
	domain "go-ddd-book-manager/internal/domain/user"
	ic "go-ddd-book-manager/internal/infra/common"
)

func userModelFromDomain(u domain.User) model {
	return model{
		Audit: ic.AuditFromDomain(u.Audit()),
		Name:  u.Name(),
		Email: u.Email(),
		Role:  u.Role(),
	}
}

func restoreUser(m model) (domain.User, error) {
	audit, err := m.Audit.Restore()
	if err != nil {
		return domain.User{}, err
	}
	return domain.Restore(audit, m.Name, m.Email, m.Role)
}
