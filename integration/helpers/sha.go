package helpers

import (
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
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

// Sha256Sum calculates the SHA256 sum of a file.
func Sha256Sum(path string) string {
	f, err := os.Open(path)
	Expect(err).ToNot(HaveOccurred())
	defer f.Close()

	hash := sha256.New()
	_, err = io.Copy(hash, f)
	Expect(err).ToNot(HaveOccurred())

	return fmt.Sprintf("%x", hash.Sum(nil))
}

// Sha256SumDirectory calculates the SHA256 sum of a directory.
func Sha256SumDirectory(dir string) string {
	tempZipFile, err := ioutil.TempFile(os.TempDir(), "")
	Expect(err).ToNot(HaveOccurred())

	err = Zipit(dir, tempZipFile.Name(), "")
	Expect(err).ToNot(HaveOccurred())

	return Sha256Sum(tempZipFile.Name())
}
