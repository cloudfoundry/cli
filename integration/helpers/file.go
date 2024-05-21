package helpers

import (
	"os"
	"path/filepath"
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
	tempFile, err := os.CreateTemp("", "*")
	Expect(err).NotTo(HaveOccurred())
	defer tempFile.Close()

	bytes := []byte(contents)
	_, err = tempFile.Write(bytes)
	Expect(err).NotTo(HaveOccurred())

	return tempFile.Name()
}

// TempDirAbsolutePath wraps `os.MkdirTemp`, ensuring symlinks are expanded
// before returning the path
func TempDirAbsolutePath(dir string, prefix string) string {
	tempDir, err := os.MkdirTemp(dir, prefix)
	Expect(err).NotTo(HaveOccurred())

	tempDir, err = filepath.EvalSymlinks(tempDir)
	Expect(err).NotTo(HaveOccurred())

	return tempDir
}

// TempFileAbsolutePath wraps `os.CreateTemp`, ensuring symlinks are expanded
// before returning the path
func TempFileAbsolutePath(dir string, pattern string) *os.File {
	var (
		err         error
		absoluteDir string
	)
	if dir == "" {
		absoluteDir = os.TempDir()
		absoluteDir, err = filepath.EvalSymlinks(absoluteDir)
		Expect(err).NotTo(HaveOccurred())
	} else {
		absoluteDir, err = filepath.EvalSymlinks(dir)
		Expect(err).NotTo(HaveOccurred())
	}
	tempFile, err := os.CreateTemp(absoluteDir, pattern)
	Expect(err).NotTo(HaveOccurred())

	return tempFile
}
