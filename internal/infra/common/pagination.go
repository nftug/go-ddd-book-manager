package common

import ac "go-ddd-book-manager/internal/application/common"

// PaginationOffset returns the offset for normalized pagination.
func PaginationOffset(p ac.Pagination) int64 {
	return (p.Page - 1) * p.PageSize
}
