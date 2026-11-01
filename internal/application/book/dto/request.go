package dto

import "uuid"

type BookRequest struct {
	Title       string   `json:"title"`
	Authors     []string `json:"authors"`
	ISBN        *string  `json:"isbn"`
	Description *string  `json:"description"`
} // @name BookRequest

type OwnerChangeRequest struct {
	OwnerID uuid.UUID `json:"ownerId" format:"uuid"`
} // @name OwnerChangeRequest
