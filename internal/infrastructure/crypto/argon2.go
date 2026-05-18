package crypto

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// VerifyArgon2id checks password against a PHC-format argon2id hash.
// Uses constant-time comparison to prevent timing attacks.
func VerifyArgon2id(password, phc string) (bool, error) {
	parts := strings.Split(phc, "$")
	// PHC format: $argon2id$v=19$m=65536,t=1,p=4$<salt>$<hash> → 6 parts after split
	if len(parts) != 6 {
		return false, fmt.Errorf("crypto: invalid PHC string")
	}

	var m, t uint32
	var p uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &m, &t, &p); err != nil {
		return false, fmt.Errorf("crypto: invalid PHC params: %w", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("crypto: invalid PHC salt: %w", err)
	}

	expected, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("crypto: invalid PHC hash: %w", err)
	}

	computed := argon2.IDKey([]byte(password), salt, t, m, p, uint32(len(expected)))
	return subtle.ConstantTimeCompare(computed, expected) == 1, nil
}

// HashArgon2id hashes a password using Argon2id and returns a PHC-format string.
// Parameters: time=1, memory=64MB, threads=4, keyLen=32
func HashArgon2id(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, 1, 65536, 4, 32)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)
	return fmt.Sprintf("$argon2id$v=%d$m=65536,t=1,p=4$%s$%s", argon2.Version, b64Salt, b64Hash), nil
}
