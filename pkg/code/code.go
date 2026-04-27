// Package code generates random invite codes.
package code

import (
	"crypto/rand"
	"encoding/base32"
)

// Generate returns an 8-character base32 (no padding) random string.
func Generate() (string, error) {
	b := make([]byte, 5) // 5 bytes → 8 base32 chars
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b), nil
}
