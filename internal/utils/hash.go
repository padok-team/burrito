package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

// ShortHash returns a short, deterministic, label-safe hash of value: 16
// lowercase hex characters, well within Kubernetes' 63-byte label value limit
// even after adding a short prefix.
func ShortHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])[:16]
}
