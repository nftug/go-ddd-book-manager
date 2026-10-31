package book

import (
	"testing"
	"uuid"

	ac "go-ddd-book-manager/internal/application/common"
	dc "go-ddd-book-manager/internal/domain/common"

	"github.com/stretchr/testify/require"
)

func TestBookResponseFromModelPermissions(t *testing.T) {
	ownerID := uuid.NewV7()
	book := model{OwnerID: ownerID}
	tests := []struct {
		name  string
		actor dc.Actor
		want  ac.PermissionResponse
	}{
		{
			"owner",
			dc.NewActor(ownerID, "owner", dc.UserRole),
			ac.PermissionResponse{CanUpdate: true, CanDelete: true},
		},
		{"other user", dc.NewActor(uuid.NewV7(), "other", dc.UserRole), ac.PermissionResponse{}},
		{
			"admin",
			dc.NewActor(uuid.NewV7(), "admin", dc.AdminRole),
			ac.PermissionResponse{CanUpdate: true, CanDelete: true},
		},
		{"system", dc.NewActor(ownerID, "system", dc.SystemRole), ac.PermissionResponse{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := book.toResponse(tt.actor)
			require.Equal(t, tt.want, response.Permission)
		})
	}
}
