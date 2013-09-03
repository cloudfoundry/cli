package cf

import (
	"archive/zip"
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Zipper interface {
	Zip(dirToZip string) (zip *bytes.Buffer, err error)
}

type ApplicationZipper struct{}

func (zipper ApplicationZipper) Zip(dirOrZipFile string) (zipBuffer *bytes.Buffer, err error) {
	if strings.HasSuffix(dirOrZipFile, ".zip") {
		return readZipFile(dirOrZipFile)
	}

	return createZipFile(dirOrZipFile)
}

func readZipFile(file string) (zipBuffer *bytes.Buffer, err error) {
	var zipBytes []byte
	zipBytes, err = ioutil.ReadFile(file)
	zipBuffer = bytes.NewBuffer(zipBytes)
	return
}

func createZipFile(dir string) (zipBuffer *bytes.Buffer, err error) {
	zipBuffer = new(bytes.Buffer)
	writer := zip.NewWriter(zipBuffer)
	exclusions := readCfIgnore(dir)

	addFileToZip := func(path string, f os.FileInfo, inErr error) (err error) {
		err = inErr
		if err != nil {
			return
		}

		if f.IsDir() {
			return
		}

		fileName := strings.TrimPrefix(path, dir+"/")
		if fileShouldBeIgnored(exclusions, fileName) {
			return
		}

		zipFile, err := writer.Create(fileName)
		if err != nil {
			return
		}

		content, err := ioutil.ReadFile(path)
		if err != nil {
			return
		}

		_, err = zipFile.Write(content)
		if err != nil {
			return
		}

		return
	}

	err = filepath.Walk(dir, addFileToZip)

	if err != nil {
		return
	}

	err = writer.Close()
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
	cfIgnore, err := os.Open(dir + "/.cfignore")
	if err != nil {
		return
	}

	ignores := strings.Split(readFile(cfIgnore), "\n")
	ignores = append([]string{".cfignore"}, ignores...)

	for _, pattern := range ignores {
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
			absolutePaths, _ = filepath.Glob(dir + "/*")
		} else {
			absolutePaths, _ = filepath.Glob(dir + "/" + pattern)
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
