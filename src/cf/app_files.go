package cf

import (
	"cf/models"
	"crypto/sha1"
	"fileutils"
	"fmt"
	"glob"
	"os"
	"path"
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

type Globs []glob.Glob

func AppFilesInDir(dir string) (appFiles []models.AppFileFields, err error) {
	dir, err = filepath.Abs(dir)
	if err != nil {
		return
	}

	err = WalkAppFiles(dir, func(fileName string, fullPath string) (err error) {
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

		appFiles = append(appFiles, models.AppFileFields{
			Path: filepath.ToSlash(fileName),
			Sha1: sha1,
			Size: size,
		})

		return
	})
	return
}

func CopyFiles(appFiles []models.AppFileFields, fromDir, toDir string) (err error) {
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

func CountFiles(directory string) uint64 {
	var count uint64
	WalkAppFiles(directory, func(_, _ string) error {
		count++
		return nil
	})
	return count
}

type walkAppFileFunc func(fileName, fullPath string) (err error)

func WalkAppFiles(dir string, onEachFile walkAppFileFunc) (err error) {
	exclusions := readCfIgnoreExclusions(dir)
	inclusions := readCfIgnoreInclusions(dir)
	walkFunc := func(fullPath string, f os.FileInfo, inErr error) (err error) {
		err = inErr
		if err != nil {
			return
		}

		if f.IsDir() {
			return
		}

		if !f.Mode().IsRegular() {
			return
		}

		fileRelativePath, _ := filepath.Rel(dir, fullPath)
		fileRelativeUnixPath := filepath.ToSlash(fileRelativePath)
		if fileShouldBeIgnored(exclusions, inclusions, fileRelativeUnixPath) {
			return
		}

		err = onEachFile(fileRelativePath, fullPath)

		return
	}

	err = filepath.Walk(dir, walkFunc)
	return
}

func fileShouldBeIgnored(exclusions Globs, inclusions Globs, relPath string) bool {
	for _, inclusion := range inclusions {
		if strings.HasPrefix(inclusion.Pattern, "/") && !strings.HasPrefix(relPath, "/") {
			relPath = "/" + relPath
		}

		if inclusion.Match(relPath) {
			return false
		}
	}
	for _, exclusion := range exclusions {
		if strings.HasPrefix(exclusion.Pattern, "/") && !strings.HasPrefix(relPath, "/") {
			relPath = "/" + relPath
		}

		if exclusion.Match(relPath) {
			return true
		}
	}
	return false
}

func readCfIgnoreExclusions(dir string) (exclusions Globs) {
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
		// Lines commencing with ! describe explicit inclusion patterns
		if strings.Index(pattern, "!") == 0 {
			continue
		}

		pattern = path.Clean(pattern)
		patternExclusions := globsForPattern(pattern)
		exclusions = append(exclusions, patternExclusions...)
	}
	return
}

func readCfIgnoreInclusions(dir string) (inclusions Globs) {
	cfIgnore, err := os.Open(filepath.Join(dir, ".cfignore"))
	if err != nil {
		return
	}

	includes := strings.Split(fileutils.ReadFile(cfIgnore), "\n")
	for _, pattern := range includes {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}
		// Lines commencing with ! describe explicit inclusion patterns
		if strings.Index(pattern, "!") != 0 {
			continue
		}
		pattern := pattern[1:]
		pattern = path.Clean(pattern)
		patternInclusions := globsForPattern(pattern)
		inclusions = append(inclusions, patternInclusions...)
	}

	return
}

func globsForPattern(cfignorePattern string) (globs Globs) {
	globs = append(globs, glob.MustCompileGlob(cfignorePattern))
	globs = append(globs, glob.MustCompileGlob(path.Join(cfignorePattern, "*")))
	globs = append(globs, glob.MustCompileGlob(path.Join(cfignorePattern, "**", "*")))

	if !strings.HasPrefix(cfignorePattern, "/") {
		globs = append(globs, glob.MustCompileGlob(path.Join("**", cfignorePattern)))
		globs = append(globs, glob.MustCompileGlob(path.Join("**", cfignorePattern, "*")))
		globs = append(globs, glob.MustCompileGlob(path.Join("**", cfignorePattern, "**", "*")))
	}
	return
}
