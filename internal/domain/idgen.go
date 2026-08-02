package domain

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// NewID returns a prefixed opaque identifier using crypto/rand.
func NewID(prefix string) (ID, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("generate id: %w", err)
	}
	return ID(prefix + hex.EncodeToString(b[:])), nil
}
