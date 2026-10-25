package common

type PermissionPolicy interface {
	CanCreate() bool
	CanUpdate() bool
	CanDelete() bool
}

func EnsureCanCreate(permission PermissionPolicy) error {
	if !permission.CanCreate() {
		return ErrForbidden
	}
	return nil
}

func EnsureCanUpdate(permission PermissionPolicy) error {
	if !permission.CanUpdate() {
		return ErrForbidden
	}
	return nil
}

func EnsureCanDelete(permission PermissionPolicy) error {
	if !permission.CanDelete() {
		return ErrForbidden
	}
	return nil
}
