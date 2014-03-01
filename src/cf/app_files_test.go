package cf_test

import (
	"path/filepath"

	. "cf"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AppFiles", func() {

	Context("AppFilesInDir", func() {
		FIt("all files have '/' path separators", func() {
			files, err := AppFilesInDir(filepath.Join("..", "fixtures", "applications"))
			Ω(err).ShouldNot(HaveOccurred())

			for _, afile := range files {
				Ω(afile.Path).Should(Equal(filepath.ToSlash(afile.Path)))
			}
		})
	})

})
