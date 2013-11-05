package cf

import (
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"io"
	"math"
	"math/big"
	"os"
	"path/filepath"
	"time"
)

func AppFilesInDir(dir string) (appFiles []AppFile, err error) {
	err = walkAppFiles(dir, func(fileName string, fullPath string) {
		fileInfo, err := os.Lstat(fullPath)
		if err != nil {
			return
		}
		size := fileInfo.Size()

		h := sha1.New()
		file, err := os.Open(fullPath)
		if err != nil {
			return
		}

		_, err = io.Copy(h, file)
		if err != nil {
			return
		}

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

func TempDirForApp() (appDir string, err error) {
	dir, err := baseTempDir()
	if err != nil {
		return
	}
	appDir = filepath.Join(dir, "apps", uniqueKey())
	return
}

func TempFileForZip() (file string, err error) {
	dir, err := baseTempDir()
	if err != nil {
		return
	}

	fileName := fmt.Sprintf("%s.zip", uniqueKey())

	file = filepath.Join(dir, "uploads", fileName)
	return
}

func TempFileForRequestBody() (file string, err error) {
	dir, err := baseTempDir()
	if err != nil {
		return
	}

	fileName := fmt.Sprintf("%s.txt", uniqueKey())

	file = filepath.Join(dir, "requests", fileName)
	return
}

func baseTempDir() (dir string, err error) {
	dir = filepath.Join(os.TempDir(), "cf")
	err = os.MkdirAll(dir, os.ModeDir|os.ModeTemporary|os.ModePerm)
	return
}

// uniqueKey creates one key per execution of the CLI

var cachedUniqueKey string

func uniqueKey() string {
	if cachedUniqueKey == "" {
		salt, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt32))
		if err != nil {
			salt = big.NewInt(1)
		}

		cachedUniqueKey = fmt.Sprintf("%d_%d", time.Now().Unix(), salt)
	}

	return cachedUniqueKey
}

func InitializeDir(dir string) (err error) {
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
