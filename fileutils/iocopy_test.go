// +build darwin freebsd linux netbsd openbsd

package fileutils_test

import (
	. "github.com/cloudfoundry/cli/fileutils"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Iocopy", func() {
	Describe(".CopyFile", func() {
		It("copies a file with correct permisions", func() {
			someFile, err := ioutil.TempFile("", "blarg")
			err = os.Chmod(someFile.Name(), 0731)
			Expect(err).ToNot(HaveOccurred())

			newDir, _ := ioutil.TempDir("", "")

			err = CopyFile(filepath.Join(newDir, "baz"), someFile.Name())
			Expect(err).ToNot(HaveOccurred())

			fileStat, _ := os.Stat(filepath.Join(newDir, "baz"))
			Expect(int(fileStat.Mode())).To(Equal(0731))
		})
	})
})
