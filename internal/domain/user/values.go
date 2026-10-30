package user

import (
	"net/mail"
	"unicode/utf8"

	"go-ddd-book-manager/internal/domain/common"
)

type Name struct{ value string }

func NewName(value string) (Name, error) {
	if value == "" {
		return Name{}, common.NewValidationError("user.name.empty")
	}
	if utf8.RuneCountInString(value) > 100 {
		return Name{}, common.NewValidationError("user.name.too_long")
	}
	return Name{value}, nil
}
func RestoreName(value string) Name { return Name{value} }
func (n Name) Value() string        { return n.value }

type Email struct{ value string }

func NewEmail(value string) (Email, error) {
	validEmail := func(value string) bool {
		if len(value) > 254 {
			return false
		}
		address, err := mail.ParseAddress(value)
		return err == nil && address.Address == value
	}
	if !validEmail(value) {
		return Email{}, common.NewValidationError("user.email.invalid")
	}
	return Email{value}, nil
}
func RestoreEmail(value string) Email { return Email{value} }
func (e Email) Value() string         { return e.value }
