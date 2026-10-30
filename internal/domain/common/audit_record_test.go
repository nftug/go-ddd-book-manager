package common_test

import (
	"testing"
	"time"

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
		if err != nil {
			t.Fatal(err)
		}
		if got.IsSome() {
			t.Fatal("all nil options should return None")
		}
	})

	t.Run("all non-nil returns some", func(t *testing.T) {
		got, err := common.RestoreAuditRecordOption(
			option.Some(id), option.Some("Aki"), option.Some(timestamp),
		)
		if err != nil {
			t.Fatal(err)
		}
		record, ok := got.Get()
		if !ok {
			t.Fatal("all non-nil options should return Some")
		}
		if record.UserID() != id || record.UserName() != "Aki" || !record.Timestamp().Equal(timestamp) {
			t.Fatalf("restored audit record = %#v, want supplied fields", record)
		}
	})

	t.Run("partially nil returns error", func(t *testing.T) {
		got, err := common.RestoreAuditRecordOption(
			option.Some(id), option.None[string](), option.Some(timestamp),
		)
		if err == nil {
			t.Fatal("partially nil options should return an error")
		}
		if got.IsSome() {
			t.Fatal("partially nil options should not return Some")
		}
	})
}
