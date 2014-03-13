package app_files

import (
	"cf/models"
	"crypto/sha1"
	"fileutils"
	"fmt"
	"os"
	"path/filepath"
)

var DefaultIgnoreFiles = []string{
	".cfignore",
	".gitignore",
	".git",
	".svn",
	"_darcs",
}

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
		err = fileutils.CopyPathToPath(fromPath, toPath)
		if err != nil {
			return
		}

		fileutils.SetModeFromPath(toPath, fromPath)
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
	cfIgnore := loadIgnoreFile(dir)
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

		if !cfIgnore.FileShouldBeIgnored(fileRelativeUnixPath) {
			err = onEachFile(fileRelativePath, fullPath)
		}

		return
	}

	err = filepath.Walk(dir, walkFunc)
	return
}

func loadIgnoreFile(dir string) CfIgnore {
	file, err := os.Open(filepath.Join(dir, ".cfignore"))
	if err == nil {
		return NewCfIgnore(fileutils.Read(file))
	} else {
		return NewCfIgnore("")
	}
}
