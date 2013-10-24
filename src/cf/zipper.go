package cf

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Zipper interface {
	Zip(dirToZip string) (zipFile *os.File, err error)
}

type ApplicationZipper struct{}

var doNotZipExtensions = []string{".zip", ".war", ".jar"}

func (zipper ApplicationZipper) Zip(dirOrZipFile string) (zipFile *os.File, err error) {
	if shouldNotZip(filepath.Ext(dirOrZipFile)) {
		return os.Open(dirOrZipFile)
	}

	zipFile, err = createZipFile(dirOrZipFile)
	if err != nil {
		return
	}

	_, err = zipFile.Seek(0, os.SEEK_SET)
	return
}

func shouldNotZip(extension string) (result bool) {
	for _, ext := range doNotZipExtensions {
		if ext == extension {
			return true
		}
	}
	return
}

func createZipFile(dir string) (zipFile *os.File, err error) {
	zipFile, err = os.Create(TempFileForZip())
	if err != nil {
		return
	}

	isEmpty, err := IsDirEmpty(dir)
	if err != nil {
		return
	}
	if isEmpty {
		err = errors.New("Directory is empty")
		return
	}

	writer := zip.NewWriter(zipFile)
	defer writer.Close()

	err = walkAppFiles(dir, func(fileName string, fullPath string) {
		zipFile, err := writer.Create(fileName)
		if err != nil {
			return
		}

		file, err := os.Open(fullPath)
		if err != nil {
			return
		}

		_, err = io.Copy(zipFile, file)
		if err != nil {
			return
		}

		return
	})

	return
}

func fileShouldBeIgnored(exclusions []string, relativePath string) bool {
	for _, exclusion := range exclusions {
		if exclusion == relativePath {
			return true
		}
	}
	return false
}

func readCfIgnore(dir string) (exclusions []string) {
	cfIgnore, err := os.Open(filepath.Join(dir, ".cfignore"))
	if err != nil {
		return
	}

	ignores := strings.Split(readFile(cfIgnore), "\n")
	ignores = append([]string{".cfignore"}, ignores...)

	for _, pattern := range ignores {
		pattern = filepath.Clean(pattern)

		patternExclusions := exclusionsForPattern(dir, pattern)
		exclusions = append(exclusions, patternExclusions...)
	}

	return
}

func exclusionsForPattern(dir string, pattern string) (exclusions []string) {
	starting_dir := dir

	findPatternMatches := func(dir string, f os.FileInfo, inErr error) (err error) {
		err = inErr
		if err != nil {
			return
		}

		absolutePaths := []string{}
		if f.IsDir() && f.Name() == pattern {
			absolutePaths, _ = filepath.Glob(filepath.Join(dir, "*"))
		} else {
			absolutePaths, _ = filepath.Glob(filepath.Join(dir, pattern))
		}

		for _, p := range absolutePaths {
			relpath, _ := filepath.Rel(starting_dir, p)

			exclusions = append(exclusions, relpath)
		}
		return
	}

	err := filepath.Walk(dir, findPatternMatches)
	if err != nil {
		return
	}

	return
}

func readFile(file *os.File) string {
	buf := &bytes.Buffer{}
	_, err := io.Copy(buf, file)

	if err != nil {
		return ""
	}

	return string(buf.Bytes())
}
