package cf

import (
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

func AppFilesInDir(dir string) (appFiles []AppFile, err error) {
	err = walkAppFiles(dir, func(fileName string, fullPath string) {
		fileInfo, err := os.Lstat(fullPath)
		if err != nil {
			return
		}
		size := fileInfo.Size()

		h := sha1.New()
		fileBytes, err := ioutil.ReadFile(fullPath)
		if err != nil {
			return
		}

		h.Write(fileBytes)
		sha1Bytes := h.Sum(nil)
		sha1 := fmt.Sprintf("%x", sha1Bytes)

		appFiles = append(appFiles, AppFile{
			Path: fileName,
			Sha1: sha1,
			Size: size,
		})
	})
	return
}

func TempDirForApp(app Application) (dir string) {
	dir = filepath.Join(os.TempDir(), "cf", app.Guid)
	return
}

func InitializeDir(dir string) (err error) {
	err = os.RemoveAll(dir)
	if err != nil {
		return
	}
	err = os.MkdirAll(dir, os.ModeDir|os.ModeTemporary|os.ModePerm)
	return
}

func CopyFiles(appFiles []AppFile, fromDir, toDir string) (err error) {
	if err != nil {
		return
	}

	for _, file := range appFiles {
		fromPath := filepath.Join(fromDir, file.Path)
		toPath := filepath.Join(toDir, file.Path)
		err = copyFile(fromPath, toPath)
		if err != nil {
			return
		}
	}
	return
}

func copyFile(fromPath, toPath string) (err error) {
	err = os.MkdirAll(filepath.Dir(toPath), os.ModeDir|os.ModeTemporary|os.ModePerm)
	if err != nil {
		return
	}

	src, err := os.Open(fromPath)
	if err != nil {
		return
	}
	defer src.Close()

	dst, err := os.Create(toPath)
	if err != nil {
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return
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

type walkAppFileFunc func(fileName, fullPath string)

func walkAppFiles(dir string, onEachFile walkAppFileFunc) (err error) {
	exclusions := readCfIgnore(dir)

	walkFunc := func(fullPath string, f os.FileInfo, inErr error) (err error) {
		err = inErr
		if err != nil {
			return
		}

		if f.IsDir() {
			return
		}

		fileName, _ := filepath.Rel(dir, fullPath)
		if fileShouldBeIgnored(exclusions, fileName) {
			return
		}

		onEachFile(fileName, fullPath)

		return
	}

	err = filepath.Walk(dir, walkFunc)
	return
}
