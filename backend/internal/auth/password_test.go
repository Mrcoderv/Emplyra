package auth

import (
	"strings"
	"testing"

	"github.com/emplyra/backend/internal/models"
)

func TestHashAndVerify(t *testing.T) {
	hash, err := HashPassword("Hunter2$ecret")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if hash == "Hunter2$ecret" {
		t.Fatal("hash must not equal plaintext")
	}
	if !VerifyPassword(hash, "Hunter2$ecret") {
		t.Fatal("expected correct password to verify")
	}
	if VerifyPassword(hash, "wrong") {
		t.Fatal("expected wrong password to fail verification")
	}
}

func TestRefreshTokenHashing(t *testing.T) {
	raw := GenerateRefreshToken()
	if len(raw) < 40 {
		t.Fatalf("refresh token too short: %d", len(raw))
	}
	h1 := HashRefreshToken(raw)
	h2 := HashRefreshToken(raw)
	if h1 != h2 {
		t.Fatal("hashing must be stable")
	}
	if HashRefreshToken(raw) == HashRefreshToken("other") {
		t.Fatal("different tokens must hash differently")
	}
}

func TestNormalizeEmail(t *testing.T) {
	if got := NormalizeEmail("  John.Doe@Example.COM  "); got != "john.doe@example.com" {
		t.Fatalf("normalize: got %q", got)
	}
}

func TestUserStatusAllowed(t *testing.T) {
	if !UserStatusAllowed(models.UserStatusActive) {
		t.Fatal("ACTIVE should be allowed")
	}
	for _, bad := range []models.UserStatus{"DISABLED", "LOCKED", ""} {
		if UserStatusAllowed(bad) {
			t.Fatalf("%q should not be allowed", bad)
		}
	}
}

func TestBcryptHashHasPrefix(t *testing.T) {
	hash, err := HashPassword("x")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(hash, "$2a$") && !strings.HasPrefix(hash, "$2b$") {
		t.Fatalf("unexpected bcrypt prefix: %q", hash)
	}
}
