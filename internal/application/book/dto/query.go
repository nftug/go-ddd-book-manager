package dto

import ac "go-ddd-book-manager/internal/application/common"

type BookQuery struct {
	Title      *string `query:"title"`
	Author     *string `query:"author"`
	CheckedOut *bool   `query:"checked_out"`
	Pagination ac.Pagination
}
