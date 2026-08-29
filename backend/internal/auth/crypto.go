package auth

import (
	"crypto/rand"
	"crypto/sha256"
)

func randRead(b []byte) {
	_, _ = rand.Read(b)
}

func sha256Sum(b []byte) [32]byte {
	return sha256.Sum256(b)
}
