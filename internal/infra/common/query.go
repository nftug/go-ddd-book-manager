package common

import "github.com/uptrace/bun"

// UpsertWithAudit updates the conflicting row and its audit fields.
func UpsertWithAudit(query *bun.InsertQuery) *bun.InsertQuery {
	return query.
		On("CONFLICT (id) DO UPDATE").
		Set("updated_by_id = EXCLUDED.updated_by_id").
		Set("updated_by_name = EXCLUDED.updated_by_name").
		Set("updated_at = EXCLUDED.updated_at")
}
