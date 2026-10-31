package common

import "uuid"

type CreatedResponse struct {
	ID uuid.UUID `json:"id"`
}
