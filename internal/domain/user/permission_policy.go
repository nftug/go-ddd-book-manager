package user

import "go-ddd-book-manager/internal/domain/common"

// userSyncPermission allows only the system actor to sync a user profile.
type userSyncPermission struct{ actor common.Actor }

func (p userSyncPermission) CanCreate() bool { return p.actor.IsSystem() }
func (p userSyncPermission) CanUpdate() bool { return p.actor.IsSystem() }
func (p userSyncPermission) CanDelete() bool { return p.actor.IsSystem() }
