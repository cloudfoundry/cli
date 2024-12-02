package trace

import (
	"crypto/rand"
	"encoding/hex"
	"strings"

	"github.com/google/uuid"
)

// GenerateUUIDTraceID returns a UUID v4 string with the dashes removed as a 32 lower-hex encoded string.
func GenerateUUIDTraceID() string {
	uuidV4 := uuid.New()
	return strings.ReplaceAll(uuidV4.String(), "-", "")
}

// GenerateRandomTraceID returns a random hex string of the given length.
func GenerateRandomTraceID(length int) string {
	b := make([]byte, length/2)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}

	return hex.EncodeToString(b)
}
