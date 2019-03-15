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

func TempFileWithContent(contents string) string {
	tempFile, err := ioutil.TempFile("", "*")
	defer tempFile.Close()
	Expect(err).NotTo(HaveOccurred())

	bytes := []byte(contents)
	_, err = tempFile.Write(bytes)
	Expect(err).NotTo(HaveOccurred())

	return tempFile.Name()
}
