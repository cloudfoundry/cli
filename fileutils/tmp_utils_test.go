package fileutils_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/cli/fileutils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TmpUtils", func() {
	Describe("TempDir", func() {
		It("creates a temporary directory", func() {
			var createdDir string
			var dirExisted bool

			fileutils.TempDir("test-prefix", func(tmpDir string, err error) {
				Expect(err).ToNot(HaveOccurred())
				Expect(tmpDir).ToNot(BeEmpty())

				createdDir = tmpDir

				// Verify directory exists during callback
				fileInfo, err := os.Stat(tmpDir)
				Expect(err).ToNot(HaveOccurred())
				Expect(fileInfo.IsDir()).To(BeTrue())
				dirExisted = true
			})

			Expect(dirExisted).To(BeTrue())

			// Verify directory is cleaned up after callback
			_, err := os.Stat(createdDir)
			Expect(os.IsNotExist(err)).To(BeTrue())
		})

		It("uses the provided name prefix", func() {
			fileutils.TempDir("my-custom-prefix", func(tmpDir string, err error) {
				Expect(err).ToNot(HaveOccurred())
				dirName := filepath.Base(tmpDir)
				Expect(dirName).To(ContainSubstring("my-custom-prefix"))
			})
		})

		It("allows writing files in the temporary directory", func() {
			var testFilePath string

			fileutils.TempDir("test-write", func(tmpDir string, err error) {
				Expect(err).ToNot(HaveOccurred())

				testFilePath = filepath.Join(tmpDir, "test-file.txt")
				err = ioutil.WriteFile(testFilePath, []byte("test content"), 0644)
				Expect(err).ToNot(HaveOccurred())

				// Verify file exists and has correct content
				content, err := ioutil.ReadFile(testFilePath)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(content)).To(Equal("test content"))
			})

			// Verify file is cleaned up
			_, err := os.Stat(testFilePath)
			Expect(os.IsNotExist(err)).To(BeTrue())
		})

		It("cleans up even if callback panics", func() {
			var createdDir string

			defer func() {
				// Recover from panic
				r := recover()
				Expect(r).ToNot(BeNil())

				// Verify cleanup happened
				_, err := os.Stat(createdDir)
				Expect(os.IsNotExist(err)).To(BeTrue())
			}()

			fileutils.TempDir("panic-test", func(tmpDir string, err error) {
				createdDir = tmpDir
				panic("intentional panic for testing")
			})
		})

		It("allows creating subdirectories", func() {
			fileutils.TempDir("subdir-test", func(tmpDir string, err error) {
				Expect(err).ToNot(HaveOccurred())

				subDir := filepath.Join(tmpDir, "subdir1", "subdir2")
				err = os.MkdirAll(subDir, 0755)
				Expect(err).ToNot(HaveOccurred())

				// Verify subdirectory exists
				fileInfo, err := os.Stat(subDir)
				Expect(err).ToNot(HaveOccurred())
				Expect(fileInfo.IsDir()).To(BeTrue())
			})
		})

		It("creates unique directories for multiple calls", func() {
			var dir1, dir2 string

			fileutils.TempDir("unique-test", func(tmpDir string, err error) {
				dir1 = tmpDir
			})

			fileutils.TempDir("unique-test", func(tmpDir string, err error) {
				dir2 = tmpDir
			})

			Expect(dir1).ToNot(Equal(dir2))
		})

		It("handles empty prefix", func() {
			fileutils.TempDir("", func(tmpDir string, err error) {
				Expect(err).ToNot(HaveOccurred())
				Expect(tmpDir).ToNot(BeEmpty())
			})
		})
	})

	Describe("TempFile", func() {
		It("creates a temporary file", func() {
			var createdFile string
			var fileExisted bool

			fileutils.TempFile("test-prefix", func(tmpFile *os.File, err error) {
				Expect(err).ToNot(HaveOccurred())
				Expect(tmpFile).ToNot(BeNil())

				createdFile = tmpFile.Name()

				// Verify file exists during callback
				fileInfo, err := os.Stat(createdFile)
				Expect(err).ToNot(HaveOccurred())
				Expect(fileInfo.IsDir()).To(BeFalse())
				fileExisted = true
			})

			Expect(fileExisted).To(BeTrue())

			// Verify file is cleaned up after callback
			_, err := os.Stat(createdFile)
			Expect(os.IsNotExist(err)).To(BeTrue())
		})

		It("uses the provided name prefix", func() {
			fileutils.TempFile("my-file-prefix", func(tmpFile *os.File, err error) {
				Expect(err).ToNot(HaveOccurred())
				fileName := filepath.Base(tmpFile.Name())
				Expect(fileName).To(ContainSubstring("my-file-prefix"))
			})
		})

		It("allows writing to the temporary file", func() {
			var testFilePath string

			fileutils.TempFile("write-test", func(tmpFile *os.File, err error) {
				Expect(err).ToNot(HaveOccurred())

				testFilePath = tmpFile.Name()

				// Write to file
				_, err = tmpFile.WriteString("test content")
				Expect(err).ToNot(HaveOccurred())

				// Seek to beginning and read
				_, err = tmpFile.Seek(0, 0)
				Expect(err).ToNot(HaveOccurred())

				content, err := ioutil.ReadAll(tmpFile)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(content)).To(Equal("test content"))
			})

			// Verify file is cleaned up
			_, err := os.Stat(testFilePath)
			Expect(os.IsNotExist(err)).To(BeTrue())
		})

		It("file is already opened when passed to callback", func() {
			fileutils.TempFile("open-test", func(tmpFile *os.File, err error) {
				Expect(err).ToNot(HaveOccurred())

				// Should be able to write immediately
				_, err = tmpFile.WriteString("immediate write")
				Expect(err).ToNot(HaveOccurred())
			})
		})

		It("closes file after callback", func() {
			var tmpFile *os.File

			fileutils.TempFile("close-test", func(tf *os.File, err error) {
				Expect(err).ToNot(HaveOccurred())
				tmpFile = tf
			})

			// Try to write to file after callback - should fail because file is closed
			_, err := tmpFile.WriteString("should fail")
			Expect(err).To(HaveOccurred())
		})

		It("cleans up even if callback panics", func() {
			var createdFile string

			defer func() {
				// Recover from panic
				r := recover()
				Expect(r).ToNot(BeNil())

				// Verify cleanup happened
				_, err := os.Stat(createdFile)
				Expect(os.IsNotExist(err)).To(BeTrue())
			}()

			fileutils.TempFile("panic-test", func(tmpFile *os.File, err error) {
				createdFile = tmpFile.Name()
				panic("intentional panic for testing")
			})
		})

		It("creates unique files for multiple calls", func() {
			var file1, file2 string

			fileutils.TempFile("unique-test", func(tmpFile *os.File, err error) {
				file1 = tmpFile.Name()
			})

			fileutils.TempFile("unique-test", func(tmpFile *os.File, err error) {
				file2 = tmpFile.Name()
			})

			Expect(file1).ToNot(Equal(file2))
		})

		It("handles empty prefix", func() {
			fileutils.TempFile("", func(tmpFile *os.File, err error) {
				Expect(err).ToNot(HaveOccurred())
				Expect(tmpFile).ToNot(BeNil())
				Expect(tmpFile.Name()).ToNot(BeEmpty())
			})
		})

		It("allows reading and writing multiple times", func() {
			fileutils.TempFile("rw-test", func(tmpFile *os.File, err error) {
				Expect(err).ToNot(HaveOccurred())

				// Write
				_, err = tmpFile.WriteString("first write\n")
				Expect(err).ToNot(HaveOccurred())

				// Write again
				_, err = tmpFile.WriteString("second write\n")
				Expect(err).ToNot(HaveOccurred())

				// Seek to beginning
				_, err = tmpFile.Seek(0, 0)
				Expect(err).ToNot(HaveOccurred())

				// Read all
				content, err := ioutil.ReadAll(tmpFile)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(content)).To(Equal("first write\nsecond write\n"))
			})
		})
	})
})
