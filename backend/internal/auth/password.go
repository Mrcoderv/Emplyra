package auth

import (
	"encoding/base64"
	"strings"

	"github.com/emplyra/backend/internal/models"
	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

func HashPassword(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func VerifyPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

func GenerateRefreshToken() string {
	b := make([]byte, 32)
	randRead(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func HashRefreshToken(token string) string {
	h := sha256Sum([]byte(token))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

func UserStatusAllowed(u models.UserStatus) bool {
	return u == models.UserStatusActive
}

func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
