package cf

import (
	"crypto/sha1"
	"fileutils"
	"fmt"
	"glob"
	"os"
	"path/filepath"
	"strings"
)

var DefaultIgnoreFiles = []string{
	".cfignore",
	".gitignore",
	".git",
	".svn",
	"_darcs",
}

type Globs []*glob.Glob

func AppFilesInDir(dir string) (appFiles []AppFileFields, err error) {
	dir, err = filepath.Abs(dir)
	if err != nil {
		return
	}

	err = walkAppFiles(dir, func(fileName string, fullPath string) (err error) {
		fileInfo, err := os.Lstat(fullPath)
		if err != nil {
			return
		}
		size := fileInfo.Size()

		h := sha1.New()

		err = fileutils.CopyPathToWriter(fullPath, h)
		if err != nil {
			return
		}

		sha1Bytes := h.Sum(nil)
		sha1 := fmt.Sprintf("%x", sha1Bytes)

		appFiles = append(appFiles, AppFileFields{
			Path: fileName,
			Sha1: sha1,
			Size: size,
		})

		return
	})
	return
}

func CopyFiles(appFiles []AppFileFields, fromDir, toDir string) (err error) {
	if err != nil {
		return
	}

	for _, file := range appFiles {
		fromPath := filepath.Join(fromDir, file.Path)
		toPath := filepath.Join(toDir, file.Path)
		err = fileutils.CopyFilePaths(fromPath, toPath)
		if err != nil {
			return
		}

		fileutils.SetExecutableBitsWithPaths(toPath, fromPath)
	}
	return
}

type walkAppFileFunc func(fileName, fullPath string) (err error)

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

		if fileShouldBeIgnored(exclusions, fullPath) {
			return
		}

		fileName, _ := filepath.Rel(dir, fullPath)
		err = onEachFile(fileName, fullPath)

		return
	}

	err = filepath.Walk(dir, walkFunc)
	return
}

func fileShouldBeIgnored(exclusions Globs, fullPath string) bool {
	for _, exclusion := range exclusions {
		if exclusion.Match(fullPath) {
			return true
		}
	}
	return false
}

func readCfIgnore(dir string) (exclusions Globs) {
	cfIgnore, err := os.Open(filepath.Join(dir, ".cfignore"))
	if err != nil {
		return
	}

	ignores := strings.Split(fileutils.ReadFile(cfIgnore), "\n")
	ignores = append(DefaultIgnoreFiles, ignores...)

	for _, pattern := range ignores {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}
		pattern = filepath.Clean(pattern)
		patternExclusions := exclusionsForPattern(dir, pattern)
		exclusions = append(exclusions, patternExclusions...)
	}
	return
}

func exclusionsForPattern(dir string, pattern string) (exclusions Globs) {
	exclusions = append(exclusions, glob.MustCompileGlob(filepath.Join(dir, pattern)))
	exclusions = append(exclusions, glob.MustCompileGlob(filepath.Join(dir, pattern, "*")))
	exclusions = append(exclusions, glob.MustCompileGlob(filepath.Join(dir, pattern, "**", "*")))

	if !filepath.IsAbs(pattern) {
		exclusions = append(exclusions, glob.MustCompileGlob(filepath.Join(dir, "**", pattern)))
		exclusions = append(exclusions, glob.MustCompileGlob(filepath.Join(dir, "**", pattern, "*")))
		exclusions = append(exclusions, glob.MustCompileGlob(filepath.Join(dir, "**", pattern, "**", "*")))
	}
	return
}
