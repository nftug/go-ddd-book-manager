package common_test

import (
	"testing"
	"time"
	"uuid"

	ic "go-ddd-book-manager/internal/infra/common"

	"github.com/stretchr/testify/require"
)

func TestAuditRejectsIncompleteUpdate(t *testing.T) {
	a := ic.Audit{
		ID:            uuid.NewV7(),
		CreatedByID:   uuid.NewV7(),
		CreatedByName: "name",
		CreatedAt:     time.Now(),
	}
	a.UpdatedByName = new("name")
	_, err := a.Restore()
	require.Error(t, err)
}
