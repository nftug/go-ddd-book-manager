package book_test

import (
	"errors"
	"testing"

	"go-ddd-book-manager/internal/domain/common"
)

func assertValidationCode(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatalf("got nil error, want validation code %q", want)
	}
	var validationErr *common.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("got error %T (%v), want ValidationError %q", err, err, want)
	}
	if got := validationErr.Code(); got != want {
		t.Fatalf("validation code = %q, want %q", got, want)
	}
}
