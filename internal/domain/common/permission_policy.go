package common

import "github.com/samber/lo"

type PermissionPolicy interface {
	CanCreate() bool
	CanUpdate() bool
	CanDelete() bool
}

func EnsureCanCreate(permission PermissionPolicy) error {
	return lo.Ternary(permission.CanCreate(), nil, ErrForbidden)
}
func EnsureCanUpdate(permission PermissionPolicy) error {
	return lo.Ternary(permission.CanUpdate(), nil, ErrForbidden)
}
func EnsureCanDelete(permission PermissionPolicy) error {
	return lo.Ternary(permission.CanDelete(), nil, ErrForbidden)
}
