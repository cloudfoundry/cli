package fileutils

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"os"
	"path/filepath"
	"time"
)

const dirMask = os.ModeDir | os.ModeTemporary | os.ModePerm

var tmpPathPrefix = ""

func SetTmpPathPrefix(path string){
	tmpPathPrefix = path
}

func TmpPathPrefix() string{
	return tmpPathPrefix
}

func TempDir(namePrefix string, cb func (tmpDir string, err error)) {
	baseDir, err := baseTempDir()
	if err != nil {
		return
	}

	tmpDir := filepath.Join(baseDir,uniqueKey(namePrefix))
	err = os.MkdirAll(tmpDir, dirMask)
	defer func() {
		os.RemoveAll(tmpDir)
	}()

	cb(tmpDir, err)
}

func TempFile(namePrefix string, cb func (tmpFile *os.File, err error)) {
	var (
		tmpFile     *os.File
		tmpFilepath string
		err         error
		tmpDir      string
	)

	tmpDir, err = baseTempDir()
	if err != nil {
		cb(tmpFile, err)
		return
	}

	tmpFilepath = filepath.Join(tmpDir, uniqueKey(namePrefix))
	tmpFile, err = os.Create(tmpFilepath)
	defer func() {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
	}()

	cb(tmpFile, err)
}

func baseTempDir() (dir string, err error) {
	dir = filepath.Join(os.TempDir(), TmpPathPrefix())
	err = os.MkdirAll(dir, dirMask)
	return
}

func uniqueKey(namePrefix string) string {
	salt, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt32))
	if err != nil {
		salt = big.NewInt(1)
	}

	return fmt.Sprintf("%s_%d_%d", namePrefix, time.Now().Unix(), salt)
}
