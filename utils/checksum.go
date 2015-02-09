package utils

import (
	"crypto/sha1"
	"io"
	"os"
)

func ComputeFileSha1(path string) ([]byte, error) {
	hash := sha1.New()

	f, err := os.Open(path)
	if err != nil {
		return []byte{}, err
	}
	defer f.Close()

	if _, err := io.Copy(hash, f); err != nil {
		return []byte{}, err
	}

	return hash.Sum(nil), nil
}
