package cf_test

import (
	"archive/zip"
	"bytes"
	. "cf"
	"fileutils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	"io"
	"os"
	"path/filepath"
)

func fileToString(file *os.File) string {
	bytesBuf := &bytes.Buffer{}
	_, err := io.Copy(bytesBuf, file)
	Expect(err).NotTo(HaveOccurred())

	return string(bytesBuf.Bytes())
}

var _ = Describe("Testing with ginkgo", func() {
	It("TestZipWithDirectory", func() {
		fileutils.TempFile("zip_test", func(zipFile *os.File, err error) {
			workingDir, err := os.Getwd()
			Expect(err).NotTo(HaveOccurred())

			dir := filepath.Join(workingDir, "../fixtures/zip/")
			err = os.Chmod(filepath.Join(dir, "subDir/bar.txt"), 0666)
			Expect(err).NotTo(HaveOccurred())

			zipper := ApplicationZipper{}
			err = zipper.Zip(dir, zipFile)
			Expect(err).NotTo(HaveOccurred())

			offset, err := zipFile.Seek(0, os.SEEK_CUR)
			Expect(err).NotTo(HaveOccurred())
			Expect(offset).To(Equal(int64(0)))

			fileStat, err := zipFile.Stat()
			Expect(err).NotTo(HaveOccurred())

			reader, err := zip.NewReader(zipFile, fileStat.Size())
			Expect(err).NotTo(HaveOccurred())

			readFileInZip := func(index int) (string, string) {
				buf := &bytes.Buffer{}
				file := reader.File[index]
				fReader, err := file.Open()
				_, err = io.Copy(buf, fReader)

				Expect(err).NotTo(HaveOccurred())

				return file.Name, string(buf.Bytes())
			}

			Expect(len(reader.File)).To(Equal(3))

			name, contents := readFileInZip(0)
			Expect(name).To(Equal("foo.txt"))
			Expect(contents).To(Equal("This is a simple text file."))

			name, contents = readFileInZip(1)
			Expect(name).To(Equal("subDir/bar.txt"))
			Expect(contents).To(Equal("I am in a subdirectory."))
			Expect(reader.File[1].FileInfo().Mode()).To(Equal(os.FileMode(0666)))

			name, contents = readFileInZip(2)
			Expect(name).To(Equal("subDir/otherDir/file.txt"))
			Expect(contents).To(Equal("This file should be present."))
		})
	})

	It("TestZipWithZipFile", func() {
		fileutils.TempFile("zip_test", func(zipFile *os.File, err error) {
			dir, err := os.Getwd()
			Expect(err).NotTo(HaveOccurred())

			zipper := ApplicationZipper{}
			err = zipper.Zip(filepath.Join(dir, "../fixtures/application.zip"), zipFile)
			Expect(err).NotTo(HaveOccurred())

			Expect(fileToString(zipFile)).To(Equal("This is an application zip file\n"))
		})
	})

	It("TestZipWithWarFile", func() {
		fileutils.TempFile("zip_test", func(zipFile *os.File, err error) {
			dir, err := os.Getwd()
			Expect(err).NotTo(HaveOccurred())

			zipper := ApplicationZipper{}
			err = zipper.Zip(filepath.Join(dir, "../fixtures/application.war"), zipFile)
			Expect(err).NotTo(HaveOccurred())

			Expect(fileToString(zipFile)).To(Equal("This is an application war file\n"))
		})
	})

	It("TestZipWithJarFile", func() {
		fileutils.TempFile("zip_test", func(zipFile *os.File, err error) {
			dir, err := os.Getwd()
			Expect(err).NotTo(HaveOccurred())

			zipper := ApplicationZipper{}
			err = zipper.Zip(filepath.Join(dir, "../fixtures/application.jar"), zipFile)
			Expect(err).NotTo(HaveOccurred())

			Expect(fileToString(zipFile)).To(Equal("This is an application jar file\n"))
		})
	})

	It("TestZipWithInvalidFile", func() {
		fileutils.TempFile("zip_test", func(zipFile *os.File, err error) {
			zipper := ApplicationZipper{}
			err = zipper.Zip("/a/bogus/directory", zipFile)
			Expect(err).To(HaveOccurred())
			assert.Contains(mr.T(), err.Error(), "open /a/bogus/directory")
		})
	})

	It("TestZipWithEmptyDir", func() {
		fileutils.TempFile("zip_test", func(zipFile *os.File, err error) {
			fileutils.TempDir("zip_test", func(emptyDir string, err error) {
				zipper := ApplicationZipper{}
				err = zipper.Zip(emptyDir, zipFile)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Directory is empty"))
			})
		})
	})
})
