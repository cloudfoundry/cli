package random

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateHex returns a random hex string of the given length.
func GenerateHex(length int) string {
	b := make([]byte, length/2)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}

	return hex.EncodeToString(b)
}
