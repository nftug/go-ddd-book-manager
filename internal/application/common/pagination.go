package common

import (
	"math"
)

type Pagination struct {
	Page     int64 `query:"page"`
	PageSize int64 `query:"page_size"`
}

// Offset returns the offset for normalized pagination.
func (p Pagination) Offset() int64 {
	return (p.Page - 1) * p.PageSize
}

func (p Pagination) Normalize() (Pagination, error) {
	if p.Page == 0 {
		p.Page = 1
	}
	if p.PageSize == 0 {
		p.PageSize = 10
	}
	if p.Page < 1 || p.PageSize < 1 || p.PageSize > 100 || p.Page-1 > math.MaxInt64/p.PageSize {
		return p, PaginationInvalidErr
	}
	return p, nil
}
