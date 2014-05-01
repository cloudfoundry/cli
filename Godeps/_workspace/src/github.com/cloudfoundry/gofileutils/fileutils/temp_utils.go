package fileutils

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

func TempDir(namePrefix string, cb func(tmpDir string, err error)) {
	tmpDir, err := ioutil.TempDir("", namePrefix)

	defer func() {
		os.RemoveAll(tmpDir)
	}()

	cb(tmpDir, err)
}

func TempFile(namePrefix string, cb func(tmpFile *os.File, err error)) {
	tmpFile, err := ioutil.TempFile("", namePrefix)

	defer func() {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
	}()

	cb(tmpFile, err)
}

// TempPath generates a random file path in tmp, but does
// NOT create the actual directory
func TempPath(namePrefix string) string {
	return filepath.Join(os.TempDir(), namePrefix, nextSuffix())
}

// copied from http://golang.org/src/pkg/io/ioutil/tempfile.go
// Random number state.
// We generate random temporary file names so that there's a good
// chance the file doesn't exist yet - keeps the number of tries in
// TempFile to a minimum.
var rand uint32
var randmu sync.Mutex

func reseed() uint32 {
	return uint32(time.Now().UnixNano() + int64(os.Getpid()))
}

func nextSuffix() string {
	randmu.Lock()
	r := rand
	if r == 0 {
		r = reseed()
	}
	r = r*1664525 + 1013904223 // constants from Numerical Recipes
	rand = r
	randmu.Unlock()
	return strconv.Itoa(int(1e9 + r%1e9))[1:]
}
