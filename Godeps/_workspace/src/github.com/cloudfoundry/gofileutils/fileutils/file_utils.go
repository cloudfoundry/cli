package fileutils

import (
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)

func Open(path string) (file *os.File, err error) {
	err = os.MkdirAll(filepath.Dir(path), os.ModeDir|os.ModePerm)
	if err != nil {
		return
	}

	return os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
}

func Create(path string) (file *os.File, err error) {
	err = os.MkdirAll(filepath.Dir(path), os.ModeDir|os.ModePerm)
	if err != nil {
		return
	}

	return os.Create(path)
}

func CopyPathToPath(fromPath, toPath string) (err error) {
	srcFileInfo, err := os.Stat(fromPath)
	if err != nil {
		return err
	}

	if srcFileInfo.IsDir() {
		err = os.MkdirAll(toPath, srcFileInfo.Mode())
		if err != nil {
			return err
		}

		files, err := ioutil.ReadDir(fromPath)
		if err != nil {
			return err
		}

		for _, file := range files {
			err = CopyPathToPath(path.Join(fromPath, file.Name()), path.Join(toPath, file.Name()))
			if err != nil {
				return err
			}
		}
	} else {
		var dst *os.File
		dst, err = Create(toPath)
		if err != nil {
			return err
		}
		defer dst.Close()

		dst.Chmod(srcFileInfo.Mode())

		err = CopyPathToWriter(fromPath, dst)
	}
	return err
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
	destFile, err := Create(targetPath)
	if err != nil {
		return
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, src)
	return
}
