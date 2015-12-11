package app_files

import (
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/gofileutils/fileutils"
)

type AppFiles interface {
	AppFilesInDir(dir string) (appFiles []models.AppFileFields, err error)
	CopyFiles(appFiles []models.AppFileFields, fromDir, toDir string) (err error)
	CountFiles(directory string) int64
	WalkAppFiles(dir string, onEachFile func(string, string) error) (err error)
}

type ApplicationFiles struct{}

func (appfiles ApplicationFiles) AppFilesInDir(dir string) (appFiles []models.AppFileFields, err error) {
	dir, err = filepath.Abs(dir)
	if err != nil {
		return
	}

	err = appfiles.WalkAppFiles(dir, func(fileName string, fullPath string) error {
		fileInfo, err := os.Lstat(fullPath)
		if err != nil {
			return err
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
			file, err := os.Open(fullPath)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(hash, file)
			if err != nil {
				return err
			}

			appFile.Sha1 = fmt.Sprintf("%x", hash.Sum(nil))
		}

		appFiles = append(appFiles, appFile)

		return nil
	})

	return
}

func (appfiles ApplicationFiles) CopyFiles(appFiles []models.AppFileFields, fromDir, toDir string) error {
	for _, file := range appFiles {
		err := func() error {
			fromPath := filepath.Join(fromDir, file.Path)
			srcFileInfo, err := os.Stat(fromPath)
			if err != nil {
				return err
			}

			toPath := filepath.Join(toDir, file.Path)

			if srcFileInfo.IsDir() {
				err = os.MkdirAll(toPath, srcFileInfo.Mode())
				if err != nil {
					return err
				}
				return nil
			}

			var dst *os.File
			dst, err = fileutils.Create(toPath)
			if err != nil {
				return err
			}
			defer dst.Close()

			dst.Chmod(srcFileInfo.Mode())

			src, err := os.Open(fromPath)
			if err != nil {
				return err
			}
			defer src.Close()

			_, err = io.Copy(dst, src)
			if err != nil {
				return err
			}

			return nil
		}()

		if err != nil {
			return err
		}
	}

	return nil
}

func (appfiles ApplicationFiles) CountFiles(directory string) int64 {
	var count int64
	appfiles.WalkAppFiles(directory, func(_, _ string) error {
		count++
		return nil
	})
	return count
}

func (appfiles ApplicationFiles) WalkAppFiles(dir string, onEachFile func(string, string) error) error {
	cfIgnore := loadIgnoreFile(dir)
	walkFunc := func(fullPath string, f os.FileInfo, err error) error {
		fileRelativePath, _ := filepath.Rel(dir, fullPath)
		fileRelativeUnixPath := filepath.ToSlash(fileRelativePath)

		if cfIgnore.FileShouldBeIgnored(fileRelativeUnixPath) {
			if err == nil {
				if f.IsDir() {
					return filepath.SkipDir
				}
			}

			if runtime.GOOS == "windows" {
				fi, statErr := os.Lstat(`\\?\` + fullPath)
				if statErr != nil {
					return statErr
				}

				if fi.IsDir() {
					return filepath.SkipDir
				}
			}

			return err
		}

		if err != nil {
			return err
		}

		if fullPath == dir {
			return nil
		}

		if !f.Mode().IsRegular() && !f.IsDir() {
			return nil
		}

		return onEachFile(fileRelativePath, fullPath)
	}

	return filepath.Walk(dir, walkFunc)
}

func loadIgnoreFile(dir string) CfIgnore {
	fileContents, err := ioutil.ReadFile(filepath.Join(dir, ".cfignore"))
	if err != nil {
		return NewCfIgnore("")
	}

	return NewCfIgnore(string(fileContents))
}
