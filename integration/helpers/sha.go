package helpers

import (
	"crypto/sha1"
	"fmt"
	"io"
	"os"

	. "github.com/onsi/gomega"
)

// Sha1Sum calculates the SHA1 sum of a file.
func Sha1Sum(path string) string {
	f, err := os.Open(path)
	Expect(err).ToNot(HaveOccurred())
	defer f.Close()

	hash := sha1.New()
	_, err = io.Copy(hash, f)
	Expect(err).ToNot(HaveOccurred())

	return fmt.Sprintf("%x", hash.Sum(nil))
}
