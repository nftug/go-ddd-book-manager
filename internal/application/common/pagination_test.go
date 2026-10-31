package common_test

import (
	"math"
	"testing"

	ac "go-ddd-book-manager/internal/application/common"
	dc "go-ddd-book-manager/internal/domain/common"

	"github.com/stretchr/testify/require"
)

func TestPaginationOffset(t *testing.T) {
	tests := []struct {
		name       string
		pagination ac.Pagination
		want       int64
	}{
		{"first page", ac.Pagination{Page: 1, PageSize: 10}, 0},
		{"second page", ac.Pagination{Page: 2, PageSize: 3}, 3},
		{"large valid page", ac.Pagination{Page: math.MaxInt64, PageSize: 1}, math.MaxInt64 - 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.pagination.Offset())
		})
	}
}

func TestPaginationNormalizeReturnsInvalidError(t *testing.T) {
	_, err := (ac.Pagination{Page: -1}).Normalize()
	require.ErrorIs(t, err, ac.PaginationInvalidErr)
	var validation *dc.ValidationError
	require.ErrorAs(t, err, &validation)
	require.Equal(t, "pagination", validation.Field())
	require.Equal(t, "invalid", validation.Code())
}
