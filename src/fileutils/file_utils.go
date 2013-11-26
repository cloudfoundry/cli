package fileutils

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
)

func OpenFile(path string) (file *os.File, err error){
	err = os.MkdirAll(filepath.Dir(path), os.ModeDir | os.ModeTemporary | os.ModePerm)
	if err != nil {
		return
	}

	return os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
}

func CreateFile(path string) (file *os.File, err error){
	err = os.MkdirAll(filepath.Dir(path), os.ModeDir | os.ModeTemporary | os.ModePerm)
	if err != nil {
		return
	}

	return os.Create(path)
}

func ReadFile(file *os.File) string {
	buf := &bytes.Buffer{}
	_, err := io.Copy(buf, file)

	if err != nil {
		return ""
	}

	return string(buf.Bytes())
}


func CopyFilePaths(fromPath, toPath string) (err error) {
	dst, err := CreateFile(toPath)
	if err != nil {
		return
	}
	defer dst.Close()

	fileStat, err := os.Stat(fromPath)

	if err != nil {
		return
	}

    err = dst.Chmod(fileStat.Mode())
	if err != nil {
		return
	}

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
	destFile, err := CreateFile(targetPath)
	if err != nil {
		return
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, src)
	return
}
