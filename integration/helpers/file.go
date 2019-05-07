package helpers

import (
	"io/ioutil"
	"strings"

	. "github.com/onsi/gomega"
)

// ConvertPathToRegularExpression converts a windows file path into a
// string which may be embedded in a ginkgo-compatible regular expression.
func ConvertPathToRegularExpression(path string) string {
	return strings.Replace(path, "\\", "\\\\", -1)
}

// TempFileWithContent writes a temp file with given content and return the
// file name.
func TempFileWithContent(contents string) string {
	tempFile, err := ioutil.TempFile("", "*")
	Expect(err).NotTo(HaveOccurred())
	defer tempFile.Close()

	bytes := []byte(contents)
	_, err = tempFile.Write(bytes)
	Expect(err).NotTo(HaveOccurred())

	return tempFile.Name()
}
