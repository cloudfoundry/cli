package random

import (
	"crypto/rand"
	"encoding/hex"
)

func GenerateHex(length int) string {
	b := make([]byte, length/2)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}

	return hex.EncodeToString(b)
}
