package cf_test

import (
	. "cf"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"path/filepath"
)

var _ = Describe("AppFiles", func() {

	Context("AppFilesInDir", func() {
		It("all files have '/' path separators", func() {
			files, err := AppFilesInDir(filepath.Join("..", "fixtures", "applications"))
			Expect(err).ShouldNot(HaveOccurred())

			for _, afile := range files {
				Expect(afile.Path).Should(Equal(filepath.ToSlash(afile.Path)))
			}
		})
	})

})
