package app_files_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/cli/cf/app_files"

	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/gofileutils/fileutils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AppFiles", func() {
	var appFiles app_files.ApplicationFiles
	var fixturePath string

	BeforeEach(func() {
		appFiles = app_files.ApplicationFiles{}
		fixturePath = filepath.Join("..", "..", "fixtures", "applications")
	})

	Describe("AppFilesInDir", func() {
		It("all files have '/' path separators", func() {
			files, err := appFiles.AppFilesInDir(fixturePath)
			Expect(err).ShouldNot(HaveOccurred())

			for _, afile := range files {
				Expect(afile.Path).Should(Equal(filepath.ToSlash(afile.Path)))
			}
		})

		It("excludes files based on the .cfignore file", func() {
			appPath := filepath.Join(fixturePath, "app-with-cfignore")
			files, err := appFiles.AppFilesInDir(appPath)
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

				files, err := appFiles.AppFilesInDir(tempdir)
				Expect(err).ToNot(HaveOccurred())

				sizes := []int64{}
				for _, file := range files {
					sizes = append(sizes, file.Size)
				}

				Expect(sizes).To(Equal([]int64{0}))
			})
		})
	})

	Describe("CopyFiles", func() {
		It("copies only the files specified", func() {
			copyDir := filepath.Join(fixturePath, "app-copy-test")

			filesToCopy := []models.AppFileFields{
				{Path: filepath.Join("dir1")},
				{Path: filepath.Join("dir1", "child-dir", "file2.txt")},
			}

			files := []string{}

			fileutils.TempDir("copyToDir", func(tmpDir string, err error) {
				copyErr := appFiles.CopyFiles(filesToCopy, copyDir, tmpDir)
				Expect(copyErr).ToNot(HaveOccurred())

				filepath.Walk(tmpDir, func(path string, fileInfo os.FileInfo, err error) error {
					Expect(err).ToNot(HaveOccurred())

					if !fileInfo.IsDir() {
						files = append(files, fileInfo.Name())
					}
					return nil
				})
			})

			// file2.txt is in lowest subtree, thus is walked first.
			Expect(files).To(Equal([]string{
				"file2.txt",
			}))
		})
	})

	Describe("WalkAppFiles", func() {
		var cb func(string, string) error
		var seen [][]string

		BeforeEach(func() {
			seen = [][]string{}
			cb = func(fileRelativePath, fullPath string) error {
				seen = append(seen, []string{fileRelativePath, fullPath})
				return nil
			}
		})

		It("calls the callback with the relative and absolute path for each file within the given dir", func() {
			err := appFiles.WalkAppFiles(filepath.Join(fixturePath, "app-copy-test"), cb)
			Expect(err).NotTo(HaveOccurred())
			Expect(seen).To(Equal([][]string{
				{"dir1", "../../fixtures/applications/app-copy-test/dir1"},
				{"dir1/child-dir", "../../fixtures/applications/app-copy-test/dir1/child-dir"},
				{"dir1/child-dir/file2.txt", "../../fixtures/applications/app-copy-test/dir1/child-dir/file2.txt"},
				{"dir1/child-dir/file3.txt", "../../fixtures/applications/app-copy-test/dir1/child-dir/file3.txt"},
				{"dir1/file1.txt", "../../fixtures/applications/app-copy-test/dir1/file1.txt"},
				{"dir2", "../../fixtures/applications/app-copy-test/dir2"},
				{"dir2/child-dir2", "../../fixtures/applications/app-copy-test/dir2/child-dir2"},
				{"dir2/child-dir2/grandchild-dir2", "../../fixtures/applications/app-copy-test/dir2/child-dir2/grandchild-dir2"},
				{"dir2/child-dir2/grandchild-dir2/file4.txt", "../../fixtures/applications/app-copy-test/dir2/child-dir2/grandchild-dir2/file4.txt"},
			}))
		})

		Context("when the given dir contains an untraversable dir", func() {
			var nonTraversableDirPath string

			BeforeEach(func() {
				nonTraversableDirPath = filepath.Join(fixturePath, "app-copy-test", "non-traversable-dir")

				err := os.Mkdir(nonTraversableDirPath, os.ModeDir) // non-traversable without os.ModePerm
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				err := os.Remove(nonTraversableDirPath)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns an error", func() {
				err := appFiles.WalkAppFiles(filepath.Join(fixturePath, "app-copy-test"), cb)
				Expect(err).To(HaveOccurred())
			})

			Context("when the untraversable dir is .cfignored", func() {
				var cfIgnorePath string

				BeforeEach(func() {
					cfIgnorePath = filepath.Join(fixturePath, "app-copy-test", ".cfignore")
					err := ioutil.WriteFile(cfIgnorePath, []byte("non-traversable-dir\n"), os.ModePerm)
					Expect(err).NotTo(HaveOccurred())
				})

				AfterEach(func() {
					err := os.Remove(cfIgnorePath)
					Expect(err).NotTo(HaveOccurred())
				})

				It("does not return an error", func() {
					err := appFiles.WalkAppFiles(filepath.Join(fixturePath, "app-copy-test"), cb)
					Expect(err).NotTo(HaveOccurred())
				})

				It("does not call the callback with the untraversable dir", func() {
					appFiles.WalkAppFiles(filepath.Join(fixturePath, "app-copy-test"), cb)
					Expect(seen).To(Equal([][]string{
						{"dir1", "../../fixtures/applications/app-copy-test/dir1"},
						{"dir1/child-dir", "../../fixtures/applications/app-copy-test/dir1/child-dir"},
						{"dir1/child-dir/file2.txt", "../../fixtures/applications/app-copy-test/dir1/child-dir/file2.txt"},
						{"dir1/child-dir/file3.txt", "../../fixtures/applications/app-copy-test/dir1/child-dir/file3.txt"},
						{"dir1/file1.txt", "../../fixtures/applications/app-copy-test/dir1/file1.txt"},
						{"dir2", "../../fixtures/applications/app-copy-test/dir2"},
						{"dir2/child-dir2", "../../fixtures/applications/app-copy-test/dir2/child-dir2"},
						{"dir2/child-dir2/grandchild-dir2", "../../fixtures/applications/app-copy-test/dir2/child-dir2/grandchild-dir2"},
						{"dir2/child-dir2/grandchild-dir2/file4.txt", "../../fixtures/applications/app-copy-test/dir2/child-dir2/grandchild-dir2/file4.txt"},
					}))
				})
			})
		})
	})
})
