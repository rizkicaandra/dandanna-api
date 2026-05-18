package token

import (
	"crypto/rand"
	"encoding/hex"
)

// Generate returns a cryptographically random 64-char hex string (32 bytes → 256 bits of entropy).
func Generate() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic("token: crypto/rand unavailable: " + err.Error())
	}
	return hex.EncodeToString(b)
}
