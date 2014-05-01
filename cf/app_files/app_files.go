package app_files

import (
	"crypto/sha1"
	"fmt"
	"github.com/cloudfoundry/cli/cf/models"
	cffileutils "github.com/cloudfoundry/cli/fileutils"
	"github.com/cloudfoundry/gofileutils/fileutils"
	"io/ioutil"
	"os"
	"path/filepath"
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
		err = fileutils.CopyPathToPath(fromPath, toPath)
		if err != nil {
			return
		}
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
