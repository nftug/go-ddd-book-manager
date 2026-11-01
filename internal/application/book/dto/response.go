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
	CheckedOut  bool                  `json:"checked_out"`
	Audit       ac.Audit              `json:"audit"`
	Permission  ac.PermissionResponse `json:"permission"`
} // @name BookResponse

type CheckoutResponse struct {
	ID           uuid.UUID        `json:"id" format:"uuid"`
	CheckedOutTo ac.UserReference `json:"checked_out_to"`
	CheckedOutAt time.Time        `json:"checked_out_at" format:"date-time"`
	ReturnedAt   *time.Time       `json:"returned_at" format:"date-time"`
} // @name CheckoutResponse

type BookPage struct {
	Items    []BookResponse `json:"items"`
	Total    int64          `json:"total"`
	Page     int64          `json:"page"`
	PageSize int64          `json:"page_size"`
} // @name BookPage

type CheckoutPage struct {
	Items    []CheckoutResponse `json:"items"`
	Total    int64              `json:"total"`
	Page     int64              `json:"page"`
	PageSize int64              `json:"page_size"`
} // @name CheckoutPage
