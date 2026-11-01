package dto

import (
	ac "go-ddd-book-manager/internal/application/common"
	"time"
	"uuid"
)

type BookResponse struct {
	ID          uuid.UUID             `json:"id" format:"uuid"`
	Title       string                `json:"title"`
	Authors     []string              `json:"authors"`
	ISBN        *string               `json:"isbn"`
	Description *string               `json:"description"`
	Owner       ac.UserReference      `json:"owner"`
	CheckedOut  bool                  `json:"checkedOut"`
	Audit       ac.Audit              `json:"audit"`
	Permission  ac.PermissionResponse `json:"permission"`
} // @name BookResponse

type CheckoutResponse struct {
	ID           uuid.UUID        `json:"id" format:"uuid"`
	CheckedOutTo ac.UserReference `json:"checkedOutTo"`
	CheckedOutAt time.Time        `json:"checkedOutAt" format:"date-time"`
	ReturnedAt   *time.Time       `json:"returnedAt" format:"date-time"`
} // @name CheckoutResponse

type BookPage struct {
	Items    []BookResponse `json:"items"`
	Total    int64          `json:"total"`
	Page     int64          `json:"page"`
	PageSize int64          `json:"pageSize"`
} // @name BookPage

type CheckoutPage struct {
	Items    []CheckoutResponse `json:"items"`
	Total    int64              `json:"total"`
	Page     int64              `json:"page"`
	PageSize int64              `json:"pageSize"`
} // @name CheckoutPage
