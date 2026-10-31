package user_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go-ddd-book-manager/internal/domain/user"
)

func TestNameAndEmailConstructorsAndRestoreFunctions(t *testing.T) {
	name, err := user.NewName("Aki")
	require.NoError(t, err)

	assert.Equal(t, "Aki", name.Value())
	assert.Equal(t, "Restored", user.RestoreName("Restored").Value())

	email, err := user.NewEmail("aki@example.com")
	require.NoError(t, err)

	assert.Equal(t, "aki@example.com", email.Value())
	assert.Equal(t, "restored@example.com", user.RestoreEmail("restored@example.com").Value())
}

func TestNameAndEmailValidation(t *testing.T) {
	_, emptyNameErr := user.NewName("")
	_, longNameErr := user.NewName(strings.Repeat("a", 101))
	_, invalidEmailErr := user.NewEmail("invalid")

	assert.Error(t, emptyNameErr)
	assert.Error(t, longNameErr)
	assert.Error(t, invalidEmailErr)
}
