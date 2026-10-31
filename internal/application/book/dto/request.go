package dto

import "uuid"

type BookRequest struct {
	Title       string   `json:"title"`
	Authors     []string `json:"authors"`
	ISBN        *string  `json:"isbn"`
	Description *string  `json:"description"`
}

type OwnerChangeRequest struct {
	OwnerID uuid.UUID `json:"ownerId"`
}
