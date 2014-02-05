package cf_test

import (
	"archive/zip"
	"bytes"
	. "cf"
	"fileutils"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	"io"
	"os"
	"path/filepath"
)

func fileToString(t mr.TestingT, file *os.File) string {
	bytesBuf := &bytes.Buffer{}
	_, err := io.Copy(bytesBuf, file)
	assert.NoError(t, err)

	return string(bytesBuf.Bytes())
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestZipWithDirectory", func() {
			fileutils.TempFile("zip_test", func(zipFile *os.File, err error) {
				workingDir, err := os.Getwd()
				assert.NoError(mr.T(), err)

				dir := filepath.Join(workingDir, "../fixtures/zip/")
				err = os.Chmod(filepath.Join(dir, "subDir/bar.txt"), 0666)
				assert.NoError(mr.T(), err)

				zipper := ApplicationZipper{}
				err = zipper.Zip(dir, zipFile)
				assert.NoError(mr.T(), err)

				offset, err := zipFile.Seek(0, os.SEEK_CUR)
				assert.NoError(mr.T(), err)
				assert.Equal(mr.T(), offset, 0)

				fileStat, err := zipFile.Stat()
				assert.NoError(mr.T(), err)

				reader, err := zip.NewReader(zipFile, fileStat.Size())
				assert.NoError(mr.T(), err)

				readFileInZip := func(index int) (string, string) {
					buf := &bytes.Buffer{}
					file := reader.File[index]
					fReader, err := file.Open()
					_, err = io.Copy(buf, fReader)

					assert.NoError(mr.T(), err)

					return file.Name, string(buf.Bytes())
				}

				assert.Equal(mr.T(), len(reader.File), 3)

				name, contents := readFileInZip(0)
				assert.Equal(mr.T(), name, "foo.txt")
				assert.Equal(mr.T(), contents, "This is a simple text file.")

				name, contents = readFileInZip(1)
				assert.Equal(mr.T(), name, "subDir/bar.txt")
				assert.Equal(mr.T(), contents, "I am in a subdirectory.")
				assert.Equal(mr.T(), reader.File[1].FileInfo().Mode(), uint32(0666))

				name, contents = readFileInZip(2)
				assert.Equal(mr.T(), name, "subDir/otherDir/file.txt")
				assert.Equal(mr.T(), contents, "This file should be present.")
			})
		})

		It("TestZipWithZipFile", func() {
			fileutils.TempFile("zip_test", func(zipFile *os.File, err error) {
				dir, err := os.Getwd()
				assert.NoError(mr.T(), err)

				zipper := ApplicationZipper{}
				err = zipper.Zip(filepath.Join(dir, "../fixtures/application.zip"), zipFile)
				assert.NoError(mr.T(), err)

				assert.Equal(mr.T(), fileToString(mr.T(), zipFile), "This is an application zip file\n")
			})
		})

		It("TestZipWithWarFile", func() {
			fileutils.TempFile("zip_test", func(zipFile *os.File, err error) {
				dir, err := os.Getwd()
				assert.NoError(mr.T(), err)

				zipper := ApplicationZipper{}
				err = zipper.Zip(filepath.Join(dir, "../fixtures/application.war"), zipFile)
				assert.NoError(mr.T(), err)

				assert.Equal(mr.T(), fileToString(mr.T(), zipFile), "This is an application war file\n")
			})
		})

		It("TestZipWithJarFile", func() {
			fileutils.TempFile("zip_test", func(zipFile *os.File, err error) {
				dir, err := os.Getwd()
				assert.NoError(mr.T(), err)

				zipper := ApplicationZipper{}
				err = zipper.Zip(filepath.Join(dir, "../fixtures/application.jar"), zipFile)
				assert.NoError(mr.T(), err)

				assert.Equal(mr.T(), fileToString(mr.T(), zipFile), "This is an application jar file\n")
			})
		})

		It("TestZipWithInvalidFile", func() {
			fileutils.TempFile("zip_test", func(zipFile *os.File, err error) {
				zipper := ApplicationZipper{}
				err = zipper.Zip("/a/bogus/directory", zipFile)
				assert.Error(mr.T(), err)
				assert.Contains(mr.T(), err.Error(), "open /a/bogus/directory")
			})
		})

		It("TestZipWithEmptyDir", func() {
			fileutils.TempFile("zip_test", func(zipFile *os.File, err error) {
				fileutils.TempDir("zip_test", func(emptyDir string, err error) {
					zipper := ApplicationZipper{}
					err = zipper.Zip(emptyDir, zipFile)
					assert.Error(mr.T(), err)
					assert.Equal(mr.T(), err.Error(), "Directory is empty")
				})
			})
		})
	})
}
