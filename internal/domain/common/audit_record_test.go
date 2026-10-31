package common_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go-ddd-book-manager/internal/common/option"
	"go-ddd-book-manager/internal/domain/common"
	"uuid"
)

func TestRestoreAuditRecordOption(t *testing.T) {
	id := uuid.NewV4()
	timestamp := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	t.Run("all nil returns none", func(t *testing.T) {
		got, err := common.RestoreAuditRecordOption(
			option.None[uuid.UUID](), option.None[string](), option.None[time.Time](),
		)
		require.NoError(t, err)

		assert.True(t, got.IsNone())
	})

	t.Run("all non-nil returns some", func(t *testing.T) {
		got, err := common.RestoreAuditRecordOption(
			option.Some(id), option.Some("Aki"), option.Some(timestamp),
		)
		require.NoError(t, err)

		record, ok := got.Get()
		require.True(t, ok)

		assert.Equal(t, id, record.UserID())
		assert.Equal(t, "Aki", record.UserName())
		assert.True(t, timestamp.Equal(record.Timestamp()))
	})

	t.Run("partially nil returns error", func(t *testing.T) {
		got, err := common.RestoreAuditRecordOption(
			option.Some(id), option.None[string](), option.Some(timestamp),
		)

		assert.Error(t, err)

		assert.True(t, got.IsNone())
	})
}
