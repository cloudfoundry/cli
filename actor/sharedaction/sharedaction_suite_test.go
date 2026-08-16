package sharedaction_test

import (
	"archive/zip"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSharedAction(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Shared Actions Suite")
}

var _ = BeforeSuite(func() {
	// Suppress log output during tests. This equates with setting to panic-level severity in older logging libraries
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
})

// Thanks to Svett Ralchev
// http://blog.ralch.com/tutorial/golang-working-with-zip/
func zipit(source, target, prefix string) error {
	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	if prefix != "" {
		_, err = io.WriteString(zipfile, prefix)
		if err != nil {
			return err
		}
	}

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name = strings.TrimPrefix(path, source)

		if info.IsDir() {
			header.Name += string(os.PathSeparator)
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if info.Mode().IsRegular() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(writer, file)
			return err
		} else if info.Mode()&os.ModeSymlink != 0 {
			pathInSymlink, err := os.Readlink(path)
			if err != nil {
				return err
			}
			symLinkContents := strings.NewReader(pathInSymlink)
			if _, err := io.Copy(writer, symLinkContents); err != nil {
				return err
			}
		}
		return nil
	})

	return err
}
