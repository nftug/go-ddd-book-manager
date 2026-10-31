package book_test

import (
	"testing"

	"go-ddd-book-manager/internal/domain/common"

	"github.com/stretchr/testify/require"
)

func requireValidationError(t *testing.T, err error, wantField, wantCode string) {
	t.Helper()
	var validationErr *common.ValidationError
	require.ErrorAs(t, err, &validationErr)
	require.Equal(t, wantField, validationErr.Field())
	require.Equal(t, wantCode, validationErr.Code())
}
