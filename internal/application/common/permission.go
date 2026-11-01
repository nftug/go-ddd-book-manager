package common

import dc "go-ddd-book-manager/internal/domain/common"

type PermissionResponse struct {
	CanUpdate bool `json:"canUpdate"`
	CanDelete bool `json:"canDelete"`
} // @name PermissionResponse

// PermissionFromDomain maps the domain policy to the permissions exposed by read use cases.
func PermissionFromDomain(policy dc.PermissionPolicy) PermissionResponse {
	return PermissionResponse{CanUpdate: policy.CanUpdate(), CanDelete: policy.CanDelete()}
}
