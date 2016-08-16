package appfiles_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"code.cloudfoundry.org/cli/cf/appfiles"
	"github.com/nu7hatch/gouuid"

	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/gofileutils/fileutils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type WalkAppFileArgs struct {
	relativePath string
	absolutePath string
}

func (a WalkAppFileArgs) RelativePath() string {
	return filepath.Join(strings.Split(a.relativePath, "/")...)
}

func (a WalkAppFileArgs) AbsolutePath() string {
	return filepath.Join(strings.Split(a.relativePath, "/")...)
}

func (a WalkAppFileArgs) Equal(other WalkAppFileArgs) bool {
	return a.RelativePath() == other.RelativePath() &&
		a.AbsolutePath() == other.AbsolutePath()
}

var _ = Describe("AppFiles", func() {
	var appFiles appfiles.ApplicationFiles
	var fixturePath string

	BeforeEach(func() {
		appFiles = appfiles.ApplicationFiles{}
		fixturePath = filepath.Join("..", "..", "fixtures", "applications")
	})

	Describe("AppFilesInDir", func() {
		It("all files have '/' path separators", func() {
			files, err := appFiles.AppFilesInDir(fixturePath)
			Expect(err).NotTo(HaveOccurred())

			for _, afile := range files {
				Expect(afile.Path).Should(Equal(filepath.ToSlash(afile.Path)))
			}
		})

		Context("when .cfignore is provided", func() {
			var paths []string

			BeforeEach(func() {
				appPath := filepath.Join(fixturePath, "app-with-cfignore")
				files, err := appFiles.AppFilesInDir(appPath)
				Expect(err).NotTo(HaveOccurred())

				paths = []string{}
				for _, file := range files {
					paths = append(paths, file.Path)
				}
			})

			It("excludes ignored files", func() {
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
		var actualWalkAppFileArgs []WalkAppFileArgs

		BeforeEach(func() {
			actualWalkAppFileArgs = []WalkAppFileArgs{}
			cb = func(fileRelativePath, fullPath string) error {
				actualWalkAppFileArgs = append(actualWalkAppFileArgs, WalkAppFileArgs{
					relativePath: fileRelativePath,
					absolutePath: fullPath,
				})
				return nil
			}
		})

		It("calls the callback with the relative and absolute path for each file within the given dir", func() {
			err := appFiles.WalkAppFiles(filepath.Join(fixturePath, "app-copy-test"), cb)
			Expect(err).NotTo(HaveOccurred())
			expectedArgs := []WalkAppFileArgs{
				{
					relativePath: "dir1",
					absolutePath: "../../fixtures/applications/app-copy-test/dir1",
				},
				{
					relativePath: "dir1/child-dir",
					absolutePath: "../../fixtures/applications/app-copy-test/dir1/child-dir",
				},
				{
					relativePath: "dir1/child-dir/file2.txt",
					absolutePath: "../../fixtures/applications/app-copy-test/dir1/child-dir/file2.txt",
				},
				{
					relativePath: "dir1/child-dir/file3.txt",
					absolutePath: "../../fixtures/applications/app-copy-test/dir1/child-dir/file3.txt",
				},
				{
					relativePath: "dir1/file1.txt",
					absolutePath: "../../fixtures/applications/app-copy-test/dir1/file1.txt",
				},
				{
					relativePath: "dir2",
					absolutePath: "../../fixtures/applications/app-copy-test/dir2",
				},
				{
					relativePath: "dir2/child-dir2",
					absolutePath: "../../fixtures/applications/app-copy-test/dir2/child-dir2",
				},
				{
					relativePath: "dir2/child-dir2/grandchild-dir2",
					absolutePath: "../../fixtures/applications/app-copy-test/dir2/child-dir2/grandchild-dir2",
				},
				{
					relativePath: "dir2/child-dir2/grandchild-dir2/file4.txt",
					absolutePath: "../../fixtures/applications/app-copy-test/dir2/child-dir2/grandchild-dir2/file4.txt",
				},
			}

			for i, actual := range actualWalkAppFileArgs {
				Expect(actual.Equal(expectedArgs[i])).To(BeTrue())
			}
		})

		Context("when the given dir contains an untraversable dir", func() {
			var (
				untraversableDirName string
				tmpDir               string
			)

			BeforeEach(func() {
				if runtime.GOOS == "windows" {
					Skip("This test is only for non-Windows platforms")
				}

				var err error
				tmpDir, err = ioutil.TempDir("", "untraversable-test")
				Expect(err).NotTo(HaveOccurred())

				guid, err := uuid.NewV4()
				Expect(err).NotTo(HaveOccurred())

				untraversableDirName = guid.String()
				untraversableDirPath := filepath.Join(tmpDir, untraversableDirName)

				err = os.Mkdir(untraversableDirPath, os.ModeDir) // untraversable without os.ModePerm
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				err := os.RemoveAll(tmpDir)
				Expect(err).NotTo(HaveOccurred())
			})

			Context("when the untraversable dir is .cfignored", func() {
				var cfIgnorePath string

				BeforeEach(func() {
					cfIgnorePath = filepath.Join(tmpDir, ".cfignore")
					err := ioutil.WriteFile(cfIgnorePath, []byte(untraversableDirName+"\n"), os.ModePerm)
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
					for _, actual := range actualWalkAppFileArgs {
						Expect(actual.RelativePath()).NotTo(Equal(untraversableDirName))
					}
				})
			})
		})
	})
})
