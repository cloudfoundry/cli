package app_files

import (
	"archive/zip"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/gofileutils/fileutils"
	"io"
	"os"
	"path/filepath"
)

type Zipper interface {
	Zip(dirToZip string, targetFile *os.File) (err error)
	IsZipFile(path string) bool
	Unzip(appDir string, destDir string) (err error)
	GetZipSize(zipFile *os.File) (int64, error)
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

	appfiles := ApplicationFiles{}
	return appfiles.WalkAppFiles(dir, func(fileName string, fullPath string) error {
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

func (zipper ApplicationZipper) Unzip(appDir string, destDir string) (err error) {
	r, err := zip.OpenReader(appDir)
	if err != nil {
		return
	}
	defer r.Close()

	for _, f := range r.File {
		func() {
			// Don't try to extract directories
			if f.FileInfo().IsDir() {
				return
			}

			var rc io.ReadCloser
			rc, err = f.Open()
			if err != nil {
				return
			}

			// functional scope from above is important
			// otherwise this only closes the last file handle
			defer rc.Close()

			destFilePath := filepath.Join(destDir, f.Name)

			err = fileutils.CopyReaderToPath(rc, destFilePath)
			if err != nil {
				return
			}

			err = os.Chmod(destFilePath, f.FileInfo().Mode())
			if err != nil {
				return
			}
		}()
	}

	return
}

func (zipper ApplicationZipper) GetZipSize(zipFile *os.File) (int64, error) {
	zipFileSize := int64(0)

	stat, err := zipFile.Stat()
	if err != nil {
		return 0, err
	}

	zipFileSize = int64(stat.Size())

	return zipFileSize, nil
}
