package app_files

import (
	"archive/zip"
	"bufio"
	"bytes"
	"io"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/gofileutils/fileutils"
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

func (zipper ApplicationZipper) IsZipFile(name string) bool {
	f, err := os.Open(name)
	if err != nil {
		return false
	}

	fi, err := f.Stat()
	if err != nil {
		return false
	}

	if fi.IsDir() {
		return false
	}

	_, err = zip.OpenReader(name)
	if err != nil && err == zip.ErrFormat {
		return zipper.isZipWithOffsetFileHeaderLocation(name)
	}

	return err == nil
}

func (zipper ApplicationZipper) Unzip(name string, destDir string) error {
	rc, err := zip.OpenReader(name)

	if err == nil {
		defer rc.Close()
		for _, f := range rc.File {
			err := zipper.extractFile(f, destDir)
			if err != nil {
				return err
			}
		}
	}

	if err == zip.ErrFormat {
		loc, err := zipper.zipFileHeaderLocation(name)
		if err != nil {
			return err
		}

		if loc > int64(-1) {
			f, err := os.Open(name)
			if err != nil {
				return err
			}

			defer f.Close()

			fi, err := f.Stat()
			if err != nil {
				return err
			}

			readerAt := io.NewSectionReader(f, loc, fi.Size())
			r, err := zip.NewReader(readerAt, fi.Size())
			if err != nil {
				return err
			}
			for _, f := range r.File {
				err := zipper.extractFile(f, destDir)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
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
		if err != nil {
			return err
		}

		if fileInfo.IsDir() {
			return nil
		}

		return fileutils.CopyPathToWriter(fullPath, zipFilePart)
	})
}

func (zipper ApplicationZipper) zipFileHeaderLocation(name string) (int64, error) {
	f, err := os.Open(name)
	if err != nil {
		return -1, err
	}

	defer f.Close()

	// zip file header signature, 0x04034b50, reversed due to little-endian byte order
	firstByte := byte(0x50)
	restBytes := []byte{0x4b, 0x03, 0x04}
	count := int64(-1)
	foundAt := int64(-1)

	reader := bufio.NewReader(f)

	keepGoing := true
	for keepGoing {
		count++

		b, err := reader.ReadByte()
		if err != nil {
			keepGoing = false
			break
		}

		if b == firstByte {
			nextBytes, err := reader.Peek(3)
			if err != nil {
				keepGoing = false
			}
			if bytes.Compare(nextBytes, restBytes) == 0 {
				foundAt = count
				keepGoing = false
				break
			}
		}
	}

	return foundAt, nil
}

func (zipper ApplicationZipper) isZipWithOffsetFileHeaderLocation(name string) bool {
	loc, err := zipper.zipFileHeaderLocation(name)
	if err != nil {
		return false
	}

	if loc > int64(-1) {
		f, err := os.Open(name)
		if err != nil {
			return false
		}

		defer f.Close()

		fi, err := f.Stat()
		if err != nil {
			return false
		}

		readerAt := io.NewSectionReader(f, loc, fi.Size())
		_, err = zip.NewReader(readerAt, fi.Size())
		if err == nil {
			return true
		}
	}

	return false
}

func (zipper ApplicationZipper) extractFile(f *zip.File, destDir string) error {
	if f.FileInfo().IsDir() {
		os.MkdirAll(filepath.Join(destDir, f.Name), os.ModeDir|os.ModePerm)
		return nil
	}

	var rc io.ReadCloser
	rc, err := f.Open()
	if err != nil {
		return err
	}

	defer rc.Close()

	destFilePath := filepath.Join(destDir, f.Name)

	err = fileutils.CopyReaderToPath(rc, destFilePath)
	if err != nil {
		return err
	}

	err = os.Chmod(destFilePath, f.FileInfo().Mode())
	if err != nil {
		return err
	}

	return nil
}
