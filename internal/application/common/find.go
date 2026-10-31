package common

import (
	dc "go-ddd-book-manager/internal/domain/common"
)

// FindRequired converts an absent repository result into a use-case not-found error.
func FindRequired[T any](find func() (*T, error)) (*T, error) {
	value, err := find()
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, dc.ErrNotFound
	}
	return value, nil
}
