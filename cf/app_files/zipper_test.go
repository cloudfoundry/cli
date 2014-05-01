package app_files_test

import (
	"archive/zip"
	"bytes"
	. "github.com/cloudfoundry/cli/cf/app_files"
	"github.com/cloudfoundry/gofileutils/fileutils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

func readFile(file *os.File) []byte {
	bytes, err := ioutil.ReadAll(file)
	Expect(err).NotTo(HaveOccurred())
	return bytes
}

var _ = Describe("Zipper", func() {
	It("zips directories", func() {
		fileutils.TempFile("zip_test", func(zipFile *os.File, err error) {
			workingDir, err := os.Getwd()
			Expect(err).NotTo(HaveOccurred())

			dir := filepath.Join(workingDir, "../../fixtures/zip/")
			err = os.Chmod(filepath.Join(dir, "subDir/bar.txt"), 0666)
			Expect(err).NotTo(HaveOccurred())

			zipper := ApplicationZipper{}
			err = zipper.Zip(dir, zipFile)
			Expect(err).NotTo(HaveOccurred())

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

			filenames := []string{}
			for _, file := range reader.File {
				filenames = append(filenames, file.Name)
			}

			Expect(filenames).To(Equal([]string{
				"foo.txt",
				"fooDir/",
				"fooDir/bar/",
				"lastDir/",
				"subDir/",
				"subDir/bar.txt",
				"subDir/otherDir/",
				"subDir/otherDir/file.txt",
			}))

			name, contents := readFileInZip(0)
			Expect(name).To(Equal("foo.txt"))
			Expect(contents).To(Equal("This is a simple text file."))

			name, contents = readFileInZip(5)
			Expect(name).To(Equal("subDir/bar.txt"))
			Expect(contents).To(Equal("I am in a subdirectory."))
			Expect(reader.File[5].FileInfo().Mode()).To(Equal(os.FileMode(0666)))
		})
	})

	It("is a no-op for a zipfile", func() {
		fileutils.TempFile("zip_test", func(zipFile *os.File, err error) {
			dir, err := os.Getwd()
			Expect(err).NotTo(HaveOccurred())

			zipper := ApplicationZipper{}
			fixture := filepath.Join(dir, "../../fixtures/applications/example-app.zip")
			err = zipper.Zip(fixture, zipFile)
			Expect(err).NotTo(HaveOccurred())

			zippedFile, err := os.Open(fixture)
			Expect(err).NotTo(HaveOccurred())
			Expect(readFile(zipFile)).To(Equal(readFile(zippedFile)))
		})
	})

	It("returns an error when zipping fails", func() {
		fileutils.TempFile("zip_test", func(zipFile *os.File, err error) {
			zipper := ApplicationZipper{}
			err = zipper.Zip("/a/bogus/directory", zipFile)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("open /a/bogus/directory"))
		})
	})

	It("returns an error when the directory is empty", func() {
		fileutils.TempFile("zip_test", func(zipFile *os.File, err error) {
			fileutils.TempDir("zip_test", func(emptyDir string, err error) {
				zipper := ApplicationZipper{}
				err = zipper.Zip(emptyDir, zipFile)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("is empty"))
			})
		})
	})
})
