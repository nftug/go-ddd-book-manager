package common

import (
	"uuid"

	domain "go-ddd-book-manager/internal/domain/common"

	"github.com/samber/do/v2"
)

type UUIDGenerator struct{}

func (UUIDGenerator) NewUUIDv7() uuid.UUID { return uuid.NewV7() }

func NewUUIDGenerator(do.Injector) (domain.UUIDGenerator, error) { return UUIDGenerator{}, nil }
