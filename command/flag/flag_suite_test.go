package flag_test

import (
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestFlag(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Flag Suite")
}

var tempDir string

var _ = BeforeEach(func() {
	var err error
	tempDir, err = ioutil.TempDir("", "cf-cli-")
	Expect(err).ToNot(HaveOccurred())
})

var _ = AfterEach(func() {
	err := os.RemoveAll(tempDir)
	Expect(err).ToNot(HaveOccurred())
})

func tempFile(data string) string {
	tempFile, err := ioutil.TempFile(tempDir, "")
	Expect(err).ToNot(HaveOccurred())
	_, err = tempFile.WriteString(data)
	Expect(err).ToNot(HaveOccurred())
	err = tempFile.Close()
	Expect(err).ToNot(HaveOccurred())

	return tempFile.Name()
}
