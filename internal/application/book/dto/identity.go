package dto

import "uuid"

type BookIdentity struct {
	BookID uuid.UUID `param:"book_id"`
}
