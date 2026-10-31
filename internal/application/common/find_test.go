package common_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	app "go-ddd-book-manager/internal/application/common"
	domain "go-ddd-book-manager/internal/domain/common"
)

func TestFindRequired(t *testing.T) {
	failure := errors.New("repository failed")
	value := "found"
	tests := []struct {
		name    string
		result  *string
		err     error
		wantErr error
	}{
		{name: "returns the found value", result: &value},
		{name: "maps absence to not found", wantErr: domain.ErrNotFound},
		{name: "preserves repository failure", err: failure, wantErr: failure},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := app.FindRequired(func() (*string, error) {
				return tt.result, tt.err
			})
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.Nil(t, got)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.result, got)
		})
	}
}
