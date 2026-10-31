package book

import (
	"go-ddd-book-manager/internal/domain/common"
	"uuid"
)

// PermissionsFor applies the same policy used by book mutations.
func PermissionsFor(actor common.Actor, ownerID uuid.UUID) common.PermissionPolicy {
	return bookPermission{actor: actor, ownerID: ownerID}
}
