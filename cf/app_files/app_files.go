package app_files

import (
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/cli/cf/models"
	cffileutils "github.com/cloudfoundry/cli/fileutils"
	"github.com/cloudfoundry/gofileutils/fileutils"
)

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

		appFile := models.AppFileFields{
			Path: filepath.ToSlash(fileName),
			Size: fileInfo.Size(),
		}

		if fileInfo.IsDir() {
			appFile.Sha1 = "0"
			appFile.Size = 0
		} else {
			hash := sha1.New()
			err = fileutils.CopyPathToWriter(fullPath, hash)
			if err != nil {
				return
			}
			appFile.Sha1 = fmt.Sprintf("%x", hash.Sum(nil))
		}

		appFiles = append(appFiles, appFile)
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
		err = copyPathToPath(fromPath, toPath)
		if err != nil {
			return
		}
	}
	return
}

func copyPathToPath(fromPath, toPath string) (err error) {
	srcFileInfo, err := os.Stat(fromPath)
	if err != nil {
		return
	}

	if srcFileInfo.IsDir() {
		err = os.MkdirAll(toPath, srcFileInfo.Mode())
		if err != nil {
			return
		}
	} else {
		var dst *os.File
		dst, err = fileutils.Create(toPath)
		if err != nil {
			return
		}
		defer dst.Close()

		dst.Chmod(srcFileInfo.Mode())

		err = fileutils.CopyPathToWriter(fromPath, dst)
	}
	return err
}

func CountFiles(directory string) int64 {
	var count int64
	WalkAppFiles(directory, func(_, _ string) error {
		count++
		return nil
	})
	return count
}

func WalkAppFiles(dir string, onEachFile func(string, string) error) (err error) {
	cfIgnore := loadIgnoreFile(dir)
	walkFunc := func(fullPath string, f os.FileInfo, inErr error) (err error) {
		err = inErr
		if err != nil {
			return
		}

		if fullPath == dir {
			return
		}

		if !cffileutils.IsRegular(f) && !f.IsDir() {
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
	fileContents, err := ioutil.ReadFile(filepath.Join(dir, ".cfignore"))
	if err == nil {
		return NewCfIgnore(string(fileContents))
	} else {
		return NewCfIgnore("")
	}
}
