package cf

import (
	"archive/zip"
	"bytes"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestZipWithDirectory(t *testing.T) {
	dir, err := os.Getwd()
	assert.NoError(t, err)

	zipper := ApplicationZipper{}
	zipFile, err := zipper.Zip(filepath.Join(dir, "../fixtures/zip/"))
	assert.NoError(t, err)

	byteReader := bytes.NewReader(zipFile.Bytes())
	reader, err := zip.NewReader(byteReader, int64(byteReader.Len()))
	assert.NoError(t, err)

	readFile := func(index int) (string, string) {
		buf := &bytes.Buffer{}
		file := reader.File[index]
		fReader, err := file.Open()
		_, err = io.Copy(buf, fReader)

		assert.NoError(t, err)

		return file.Name, string(buf.Bytes())
	}

	assert.Equal(t, len(reader.File), 2)

	name, contents := readFile(0)
	assert.Equal(t, name, "foo.txt")
	assert.Equal(t, contents, "This is a simple text file.")

	name, contents = readFile(1)
	assert.Equal(t, name, filepath.Clean("subDir/bar.txt"))
	assert.Equal(t, contents, "I am in a subdirectory.")
}

func TestZipWithZipFile(t *testing.T) {
	dir, err := os.Getwd()
	assert.NoError(t, err)

	zipper := ApplicationZipper{}
	zipFile, err := zipper.Zip(filepath.Join(dir, "../fixtures/application.zip"))
	assert.NoError(t, err)

	assert.Equal(t, string(zipFile.Bytes()), "This is an application zip file\n")
}

func TestZipWithWarFile(t *testing.T) {
	dir, err := os.Getwd()
	assert.NoError(t, err)

	zipper := ApplicationZipper{}
	zipFile, err := zipper.Zip(filepath.Join(dir, "../fixtures/application.war"))
	assert.NoError(t, err)

	assert.Equal(t, string(zipFile.Bytes()), "This is an application war file\n")
}

func TestZipWithJarFile(t *testing.T) {
	dir, err := os.Getwd()
	assert.NoError(t, err)

	zipper := ApplicationZipper{}
	zipFile, err := zipper.Zip(filepath.Join(dir, "../fixtures/application.jar"))
	assert.NoError(t, err)

	assert.Equal(t, string(zipFile.Bytes()), "This is an application jar file\n")
}
