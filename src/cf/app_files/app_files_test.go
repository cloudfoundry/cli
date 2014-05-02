package app_files_test

import (
	"github.com/cloudfoundry/gofileutils/fileutils"
	"os"
	"path/filepath"

	. "cf/app_files"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AppFiles", func() {
	fixturePath := filepath.Join("..", "..", "fixtures", "applications")

	Describe("AppFilesInDir", func() {
		It("all files have '/' path separators", func() {
			files, err := AppFilesInDir(fixturePath)
			Expect(err).ShouldNot(HaveOccurred())

			for _, afile := range files {
				Expect(afile.Path).Should(Equal(filepath.ToSlash(afile.Path)))
			}
		})

		It("excludes files based on the .cfignore file", func() {
			appPath := filepath.Join(fixturePath, "app-with-cfignore")
			files, err := AppFilesInDir(appPath)
			Expect(err).ShouldNot(HaveOccurred())

			paths := []string{}
			for _, file := range files {
				paths = append(paths, file.Path)
			}

			Expect(paths).To(Equal([]string{
				"dir1",
				"dir1/child-dir",
				"dir1/child-dir/file3.txt",
				"dir1/file1.txt",
				"dir2",

				// TODO: this should be excluded.
				// .cfignore doesn't handle ** patterns right now
				"dir2/child-dir2",
			}))
		})

		// NB: on windows, you can never rely on the size of a directory being zero
		// see: http://msdn.microsoft.com/en-us/library/windows/desktop/aa364946(v=vs.85).aspx
		// and: https://www.pivotaltracker.com/story/show/70470232
		It("always sets the size of directories to zero bytes", func() {
			fileutils.TempDir("something", func(tempdir string, err error) {
				Expect(err).ToNot(HaveOccurred())

				err = os.Mkdir(filepath.Join(tempdir, "nothing"), 0600)
				Expect(err).ToNot(HaveOccurred())

				files, err := AppFilesInDir(tempdir)
				Expect(err).ToNot(HaveOccurred())

				sizes := []int64{}
				for _, file := range files {
					sizes = append(sizes, file.Size)
				}

				Expect(sizes).To(Equal([]int64{0}))
			})
		})
	})
})
