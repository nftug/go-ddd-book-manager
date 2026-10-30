package user_test

import (
	"strings"
	"testing"

	"go-ddd-book-manager/internal/domain/user"
)

func TestNameAndEmailConstructorsAndRestoreFunctions(t *testing.T) {
	name, err := user.NewName("Aki")
	if err != nil || name.Value() != "Aki" {
		t.Fatalf("NewName got (%q, %v), want (Aki, nil)", name.Value(), err)
	}
	if got := user.RestoreName("Restored").Value(); got != "Restored" {
		t.Fatalf("RestoreName got %q, want Restored", got)
	}

	email, err := user.NewEmail("aki@example.com")
	if err != nil || email.Value() != "aki@example.com" {
		t.Fatalf("NewEmail got (%q, %v), want (aki@example.com, nil)", email.Value(), err)
	}
	if got := user.RestoreEmail("restored@example.com").Value(); got != "restored@example.com" {
		t.Fatalf("RestoreEmail got %q, want restored@example.com", got)
	}
}

func TestNameAndEmailValidation(t *testing.T) {
	if _, err := user.NewName(""); err == nil {
		t.Fatal("empty name should fail validation")
	}
	if _, err := user.NewName(strings.Repeat("a", 101)); err == nil {
		t.Fatal("name longer than 100 characters should fail validation")
	}
	if _, err := user.NewEmail("invalid"); err == nil {
		t.Fatal("invalid email should fail validation")
	}
}
