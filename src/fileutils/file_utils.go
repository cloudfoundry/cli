package fileutils

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
)

func CopyFilePaths(fromPath, toPath string) (err error) {
	err = os.MkdirAll(filepath.Dir(toPath), os.ModeDir | os.ModeTemporary | os.ModePerm)
	if err != nil {
		return
	}

	dst, err := os.Create(toPath)
	if err != nil {
		return
	}
	defer dst.Close()

	return CopyPathToWriter(fromPath, dst)
}

func IsDirEmpty(dir string) (isEmpty bool, err error) {
	dirFile, err := os.Open(dir)
	if err != nil {
		return
	}

	_, readErr := dirFile.Readdirnames(1)
	if readErr != nil {
		isEmpty = true
	} else {
		isEmpty = false
	}
	return
}

func CopyPathToWriter(originalFilePath string, targetWriter io.Writer) (err error) {
	originalFile, err := os.Open(originalFilePath)
	if err != nil {
		return
	}
	defer originalFile.Close()

	_, err = io.Copy(targetWriter, originalFile)
	if err != nil {
		return
	}

	return
}

func CopyReaderToPath(src io.Reader, targetPath string) (err error) {
	err = os.MkdirAll(filepath.Dir(targetPath), os.ModePerm | os.ModeDir)
	if err != nil {
		return
	}

	destFile, err := os.Create(targetPath)
	if err != nil {
		return
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, src)
	return
}

func ReadFile(file *os.File) string {
	buf := &bytes.Buffer{}
	_, err := io.Copy(buf, file)

	if err != nil {
		return ""
	}

	return string(buf.Bytes())
}
