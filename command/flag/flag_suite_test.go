package flag_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFlag(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Flag Suite")
}

var tempDir string

var _ = BeforeEach(func() {
	var err error
	tempDir, err = os.MkdirTemp("", "cf-cli-")
	Expect(err).ToNot(HaveOccurred())
})

var _ = AfterEach(func() {
	err := os.RemoveAll(tempDir)
	Expect(err).ToNot(HaveOccurred())
})

func tempFile(data string) string {
	tempFile, err := os.CreateTemp(tempDir, "")
	Expect(err).ToNot(HaveOccurred())
	_, err = tempFile.WriteString(data)
	Expect(err).ToNot(HaveOccurred())
	err = tempFile.Close()
	Expect(err).ToNot(HaveOccurred())

	return tempFile.Name()
}
