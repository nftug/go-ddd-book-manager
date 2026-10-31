package common

import (
	dc "go-ddd-book-manager/internal/domain/common"
	"math"
)

type Pagination struct {
	Page     int64 `query:"page"`
	PageSize int64 `query:"page_size"`
}

func (p Pagination) Normalize() (Pagination, error) {
	if p.Page == 0 {
		p.Page = 1
	}
	if p.PageSize == 0 {
		p.PageSize = 10
	}
	if p.Page < 1 || p.PageSize < 1 || p.PageSize > 100 || p.Page-1 > math.MaxInt64/p.PageSize {
		return p, dc.NewValidationError("pagination", "invalid")
	}
	return p, nil
}
