package fileutils_test

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"code.cloudfoundry.org/gofileutils/fileutils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Fileutils File", func() {
	var fixturePath = filepath.Clean("../fixtures/fileutils/supervirus.zsh")
	var fixtureBytes []byte

	BeforeEach(func() {
		var err error
		fixtureBytes, err = ioutil.ReadFile(fixturePath)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Open", func() {
		It("opens an existing file", func() {
			fd, err := fileutils.Open(fixturePath)
			Expect(err).NotTo(HaveOccurred())

			fileBytes, err := ioutil.ReadAll(fd)
			Expect(err).NotTo(HaveOccurred())
			fd.Close()

			Expect(fileBytes).To(Equal(fixtureBytes))
		})

		It("creates a non-existing file and all intermediary directories", func() {
			fd, err := ioutil.TempFile("", "open_test")
			Expect(err).NotTo(HaveOccurred())

			_, err = fd.WriteString("Never Gonna Give You Up")
			Expect(err).NotTo(HaveOccurred())
			fd.Close()

			fileBytes, err := ioutil.ReadFile(fd.Name())
			Expect(err).NotTo(HaveOccurred())
			Expect(string(fileBytes)).To(Equal("Never Gonna Give You Up"))
		})
	})

	Describe("Create", func() {
		It("truncates an existing file", func() {
			tmpFile, err := ioutil.TempFile("", "create_test")
			Expect(err).NotTo(HaveOccurred())
			_, err = tmpFile.WriteString("Never Gonna Give You Up")
			Expect(err).NotTo(HaveOccurred())
			filePath := tmpFile.Name()
			tmpFile.Close()

			fd, err := fileutils.Create(filePath)
			Expect(err).NotTo(HaveOccurred())

			fileBytes, err := ioutil.ReadAll(fd)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(fileBytes)).To(Equal(0))
			fd.Close()
		})

		It("creates a non-existing file and all intermediary directories", func() {
			fd, err := ioutil.TempFile("", "create_test")
			Expect(err).NotTo(HaveOccurred())

			_, err = fd.WriteString("Never Gonna Let You Down")
			Expect(err).NotTo(HaveOccurred())
			fd.Close()

			fileBytes, err := ioutil.ReadFile(fd.Name())
			Expect(err).NotTo(HaveOccurred())
			Expect(string(fileBytes)).To(Equal("Never Gonna Let You Down"))
		})
	})

	Describe("CopyPathToPath", func() {
		Describe("when the source is a file", func() {
			var destPath string

			BeforeEach(func() {
				fd, err := ioutil.TempFile("", "copy_test")
				Expect(err).NotTo(HaveOccurred())
				fd.Close()
				destPath = fd.Name()
			})

			AfterEach(func() {
				os.RemoveAll(destPath)
			})

			It("copies the file contents", func() {
				err := fileutils.CopyPathToPath(fixturePath, destPath)
				Expect(err).NotTo(HaveOccurred())
				fileBytes, err := ioutil.ReadFile(destPath)
				Expect(err).NotTo(HaveOccurred())

				fixtureBytes, err := ioutil.ReadFile(fixturePath)
				Expect(err).NotTo(HaveOccurred())
				Expect(fileBytes).To(Equal(fixtureBytes))
			})

			It("preserves the file mode", func() {
				err := fileutils.CopyPathToPath(fixturePath, destPath)
				Expect(err).NotTo(HaveOccurred())
				fileInfo, err := os.Stat(destPath)
				Expect(err).NotTo(HaveOccurred())

				expectedFileInfo, err := os.Stat(fixturePath)
				Expect(err).NotTo(HaveOccurred())

				Expect(fileInfo.Mode()).To(Equal(expectedFileInfo.Mode()))
			})
		})

		Describe("when the source is a directory", func() {
			var dirPath, destPath string

			BeforeEach(func() {
				dirPath = filepath.Join(filepath.Dir(fixturePath), "some-dir")
				destPath = filepath.Join(filepath.Dir(fixturePath), "some-other-dir")
			})

			AfterEach(func() {
				os.RemoveAll(destPath)
			})

			It("creates a directory at the destination path", func() {
				err := fileutils.CopyPathToPath(dirPath, destPath)
				Expect(err).NotTo(HaveOccurred())

				fileInfo, err := os.Stat(destPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(fileInfo.IsDir()).To(BeTrue())
			})

			It("copies all of the files from the src directory", func() {
				err := fileutils.CopyPathToPath(dirPath, destPath)
				Expect(err).NotTo(HaveOccurred())

				fileInfo, err := os.Stat(path.Join(destPath, "some-file"))
				Expect(err).NotTo(HaveOccurred())
				Expect(fileInfo.IsDir()).To(BeFalse())
			})

			It("preserves the directory's mode", func() {
				err := fileutils.CopyPathToPath(dirPath, destPath)
				Expect(err).NotTo(HaveOccurred())

				fileInfo, err := os.Stat(destPath)
				Expect(err).NotTo(HaveOccurred())

				expectedFileInfo, err := os.Stat(dirPath)
				Expect(err).NotTo(HaveOccurred())

				Expect(fileInfo.Mode()).To(Equal(expectedFileInfo.Mode()))
			})
		})
	})
})
