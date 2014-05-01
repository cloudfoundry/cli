package app_files

import (
	"archive/zip"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/gofileutils/fileutils"
	"os"
	"path/filepath"
)

type Zipper interface {
	Zip(dirToZip string, targetFile *os.File) (err error)
	IsZipFile(path string) bool
}

type ApplicationZipper struct{}

func (zipper ApplicationZipper) Zip(dirOrZipFile string, targetFile *os.File) (err error) {
	if zipper.IsZipFile(dirOrZipFile) {
		err = fileutils.CopyPathToWriter(dirOrZipFile, targetFile)
	} else {
		err = writeZipFile(dirOrZipFile, targetFile)
	}
	targetFile.Seek(0, os.SEEK_SET)
	return
}

func (zipper ApplicationZipper) IsZipFile(file string) (result bool) {
	_, err := zip.OpenReader(file)
	return err == nil
}

func writeZipFile(dir string, targetFile *os.File) error {
	isEmpty, err := fileutils.IsDirEmpty(dir)
	if err != nil {
		return err
	}

	if isEmpty {
		return errors.NewEmptyDirError(dir)
	}

	writer := zip.NewWriter(targetFile)
	defer writer.Close()

	return WalkAppFiles(dir, func(fileName string, fullPath string) error {
		fileInfo, err := os.Stat(fullPath)
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(fileInfo)
		if err != nil {
			return err
		}

		header.Name = filepath.ToSlash(fileName)

		if fileInfo.IsDir() {
			header.Name += "/"
		}

		zipFilePart, err := writer.CreateHeader(header)

		if fileInfo.IsDir() {
			return nil
		} else {
			return fileutils.CopyPathToWriter(fullPath, zipFilePart)
		}
	})
}
