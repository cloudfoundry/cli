package cf

import (
	"archive/zip"
	"errors"
	"fileutils"
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

func writeZipFile(dir string, targetFile *os.File) (err error) {
	isEmpty, err := fileutils.IsDirEmpty(dir)
	if err != nil {
		return
	}
	if isEmpty {
		err = errors.New("Directory is empty")
		return
	}

	writer := zip.NewWriter(targetFile)
	defer writer.Close()

	err = WalkAppFiles(dir, func(fileName string, fullPath string) (err error) {
		fileInfo, err := os.Stat(fullPath)
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(fileInfo)
		header.Name = filepath.ToSlash(fileName)
		if err != nil {
			return err
		}

		zipFilePart, err := writer.CreateHeader(header)
		err = fileutils.CopyPathToWriter(fullPath, zipFilePart)
		return
	})

	return
}
