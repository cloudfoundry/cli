package cf

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"os"
	"path/filepath"
	"time"
)

func TempDir(pathPrefix string, cb func(tmpDir string, err error)) {
	var (
		tmpDir string
		err    error
	)

	tmpDir, err = baseTempDir(filepath.Join(pathPrefix, uniqueKey()))
	defer func() {
		os.RemoveAll(tmpDir)
	}()

	cb(tmpDir, err)
}

func TempFile(pathPrefix string, cb func(tmpFile *os.File, err error)) {
	var (
		tmpFile     *os.File
		tmpFilepath string
		err         error
		tmpDir      string
	)

	tmpDir, err = baseTempDir(pathPrefix)
	if err != nil {
		cb(tmpFile, err)
		return
	}

	tmpFilepath = filepath.Join(tmpDir, uniqueKey())
	tmpFile, err = os.Create(tmpFilepath)
	defer func() {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
	}()

	cb(tmpFile, err)
}

func baseTempDir(subpath string) (dir string, err error) {
	dir = filepath.Join(os.TempDir(), "cf", subpath)
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
