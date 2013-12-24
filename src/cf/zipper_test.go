package cf

import (
	"archive/zip"
	"bytes"
	"fileutils"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestZipWithDirectory(t *testing.T) {
	fileutils.TempFile("zip_test", func(zipFile *os.File, err error) {

		workingDir, err := os.Getwd()
		assert.NoError(t, err)

		dir := filepath.Join(workingDir, "../fixtures/zip/")
		err = os.Chmod(filepath.Join(dir, "subDir/bar.txt"), 0666)
		assert.NoError(t, err)

		zipper := ApplicationZipper{}
		err = zipper.Zip(dir, zipFile)
		assert.NoError(t, err)

		offset, err := zipFile.Seek(0, os.SEEK_CUR)
		assert.NoError(t, err)
		assert.Equal(t, offset, 0)

		fileStat, err := zipFile.Stat()
		assert.NoError(t, err)

		reader, err := zip.NewReader(zipFile, fileStat.Size())
		assert.NoError(t, err)

		readFileInZip := func(index int) (string, string) {
			buf := &bytes.Buffer{}
			file := reader.File[index]
			fReader, err := file.Open()
			_, err = io.Copy(buf, fReader)

			assert.NoError(t, err)

			return file.Name, string(buf.Bytes())
		}

		assert.Equal(t, len(reader.File), 3)

		name, contents := readFileInZip(0)
		assert.Equal(t, name, "foo.txt")
		assert.Equal(t, contents, "This is a simple text file.")

		name, contents = readFileInZip(1)
		assert.Equal(t, name, filepath.Clean("subDir/bar.txt"))
		assert.Equal(t, contents, "I am in a subdirectory.")
		assert.Equal(t, reader.File[1].FileInfo().Mode(), uint32(0666))

		name, contents = readFileInZip(2)
		assert.Equal(t, name, filepath.Clean("subDir/otherDir/file.txt"))
		assert.Equal(t, contents, "This file should be present.")
	})
}

func TestZipWithZipFile(t *testing.T) {
	fileutils.TempFile("zip_test", func(zipFile *os.File, err error) {
		dir, err := os.Getwd()
		assert.NoError(t, err)

		zipper := ApplicationZipper{}
		err = zipper.Zip(filepath.Join(dir, "../fixtures/application.zip"), zipFile)
		assert.NoError(t, err)

		assert.Equal(t, fileToString(t, zipFile), "This is an application zip file\n")
	})
}

func TestZipWithWarFile(t *testing.T) {
	fileutils.TempFile("zip_test", func(zipFile *os.File, err error) {
		dir, err := os.Getwd()
		assert.NoError(t, err)

		zipper := ApplicationZipper{}
		err = zipper.Zip(filepath.Join(dir, "../fixtures/application.war"), zipFile)
		assert.NoError(t, err)

		assert.Equal(t, fileToString(t, zipFile), "This is an application war file\n")
	})
}

func TestZipWithJarFile(t *testing.T) {
	fileutils.TempFile("zip_test", func(zipFile *os.File, err error) {
		dir, err := os.Getwd()
		assert.NoError(t, err)

		zipper := ApplicationZipper{}
		err = zipper.Zip(filepath.Join(dir, "../fixtures/application.jar"), zipFile)
		assert.NoError(t, err)

		assert.Equal(t, fileToString(t, zipFile), "This is an application jar file\n")
	})
}

func TestZipWithInvalidFile(t *testing.T) {
	fileutils.TempFile("zip_test", func(zipFile *os.File, err error) {
		zipper := ApplicationZipper{}
		err = zipper.Zip("/a/bogus/directory", zipFile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "open /a/bogus/directory")
	})
}

func TestZipWithEmptyDir(t *testing.T) {
	fileutils.TempFile("zip_test", func(zipFile *os.File, err error) {
		fileutils.TempDir("zip_test", func(emptyDir string, err error) {
			zipper := ApplicationZipper{}
			err = zipper.Zip(emptyDir, zipFile)
			assert.Error(t, err)
			assert.Equal(t, err.Error(), "Directory is empty")
		})
	})
}

func fileToString(t *testing.T, file *os.File) string {
	bytesBuf := &bytes.Buffer{}
	_, err := io.Copy(bytesBuf, file)
	assert.NoError(t, err)

	return string(bytesBuf.Bytes())
}
