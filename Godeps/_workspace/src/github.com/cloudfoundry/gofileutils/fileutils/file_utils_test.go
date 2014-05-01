package fileutils_test

import (
	"github.com/cloudfoundry/gofileutils/fileutils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"os"
	"path/filepath"
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
			filePath := fileutils.TempPath("open_test")

			fd, err := fileutils.Open(filePath)
			Expect(err).NotTo(HaveOccurred())

			_, err = fd.WriteString("Never Gonna Give You Up")
			Expect(err).NotTo(HaveOccurred())
			fd.Close()

			fileBytes, err := ioutil.ReadFile(filePath)
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
			filePath := fileutils.TempPath("create_test")

			fd, err := fileutils.Create(filePath)
			Expect(err).NotTo(HaveOccurred())

			_, err = fd.WriteString("Never Gonna Let You Down")
			Expect(err).NotTo(HaveOccurred())
			fd.Close()

			fileBytes, err := ioutil.ReadFile(filePath)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(fileBytes)).To(Equal("Never Gonna Let You Down"))
		})
	})

	Describe("CopyPathToPath", func() {
		var destPath string

		BeforeEach(func() {
			destPath = fileutils.TempPath("copy_test")
		})

		Describe("when the source is a file", func() {
			BeforeEach(func() {
				err := fileutils.CopyPathToPath(fixturePath, destPath)
				Expect(err).NotTo(HaveOccurred())
			})

			It("copies the file contents", func() {
				fileBytes, err := ioutil.ReadFile(destPath)
				Expect(err).NotTo(HaveOccurred())

				fixtureBytes, err := ioutil.ReadFile(fixturePath)
				Expect(err).NotTo(HaveOccurred())
				Expect(fileBytes).To(Equal(fixtureBytes))
			})

			It("preserves the file mode", func() {
				fileInfo, err := os.Stat(destPath)
				Expect(err).NotTo(HaveOccurred())

				expectedFileInfo, err := os.Stat(fixturePath)
				Expect(err).NotTo(HaveOccurred())

				Expect(fileInfo.Mode()).To(Equal(expectedFileInfo.Mode()))
			})
		})

		Describe("when the source is a directory", func() {
			dirPath := filepath.Join(filepath.Dir(fixturePath), "some-dir")

			BeforeEach(func() {
				destPath = filepath.Join(destPath, "some-other-dir")
				err := fileutils.CopyPathToPath(dirPath, destPath)
				Expect(err).NotTo(HaveOccurred())
			})

			It("creates a directory at the destination path", func() {
				fileInfo, err := os.Stat(destPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(fileInfo.IsDir()).To(BeTrue())
			})

			It("preserves the directory's mode", func() {
				fileInfo, err := os.Stat(destPath)
				Expect(err).NotTo(HaveOccurred())

				expectedFileInfo, err := os.Stat(dirPath)
				Expect(err).NotTo(HaveOccurred())

				Expect(fileInfo.Mode()).To(Equal(expectedFileInfo.Mode()))
			})
		})
	})

	Describe("CopyPathToWriter", func() {
		var destPath string

		BeforeEach(func() {
			destFile, err := ioutil.TempFile("", "copy_test")
			Expect(err).NotTo(HaveOccurred())
			defer destFile.Close()

			destPath = destFile.Name()

			err = fileutils.CopyPathToWriter(fixturePath, destFile)
			Expect(err).NotTo(HaveOccurred())
		})

		It("copies the file contents", func() {
			fileBytes, err := ioutil.ReadFile(destPath)
			Expect(err).NotTo(HaveOccurred())

			Expect(fileBytes).To(Equal(fixtureBytes))
		})
	})

	Describe("CopyReaderToPath", func() {
		var destPath = fileutils.TempPath("copy_test")

		BeforeEach(func() {
			fixtureReader, err := os.Open(fixturePath)
			Expect(err).NotTo(HaveOccurred())
			defer fixtureReader.Close()

			err = fileutils.CopyReaderToPath(fixtureReader, destPath)
			Expect(err).NotTo(HaveOccurred())
		})

		It("copies the file contents", func() {
			fileBytes, err := ioutil.ReadFile(destPath)
			Expect(err).NotTo(HaveOccurred())

			Expect(fileBytes).To(Equal(fixtureBytes))
		})
	})
})
