package database_test

import (
	"testing"
	"time"
	"uuid"

	"go-ddd-book-manager/internal/infra/database"

	"github.com/stretchr/testify/require"
)

func TestAuditRejectsIncompleteUpdate(t *testing.T) {
	a := database.Audit{
		ID:            uuid.NewV7(),
		CreatedByID:   uuid.NewV7(),
		CreatedByName: "name",
		CreatedAt:     time.Now(),
	}
	a.UpdatedByName = new("name")
	_, err := a.RestoreAudit()
	require.Error(t, err)
}
