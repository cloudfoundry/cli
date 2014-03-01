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
	exclusions := readCfIgnore(dir)
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
		if fileShouldBeIgnored(exclusions, fileRelativeUnixPath) {
			return
		}

		err = onEachFile(fileRelativePath, fullPath)

		return
	}

	err = filepath.Walk(dir, walkFunc)
	return
}

func fileShouldBeIgnored(exclusions Globs, relPath string) bool {
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
		pattern = path.Clean(pattern)
		patternExclusions := exclusionsForPattern(pattern)
		exclusions = append(exclusions, patternExclusions...)
	}
	return
}

func exclusionsForPattern(cfignorePattern string) (exclusions Globs) {
	exclusions = append(exclusions, glob.MustCompileGlob(cfignorePattern))
	exclusions = append(exclusions, glob.MustCompileGlob(path.Join(cfignorePattern, "*")))
	exclusions = append(exclusions, glob.MustCompileGlob(path.Join(cfignorePattern, "**", "*")))

	if !strings.HasPrefix(cfignorePattern, "/") {
		exclusions = append(exclusions, glob.MustCompileGlob(path.Join("**", cfignorePattern)))
		exclusions = append(exclusions, glob.MustCompileGlob(path.Join("**", cfignorePattern, "*")))
		exclusions = append(exclusions, glob.MustCompileGlob(path.Join("**", cfignorePattern, "**", "*")))
	}
	return
}
