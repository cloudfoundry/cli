package ykk_test

import (
	"archive/zip"
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	. "code.cloudfoundry.org/ykk"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Reader", func() {
	Describe("NewReader", func() {
		var (
			srcDir  string
			destDir string
			zip1    string
		)

		BeforeEach(func() {
			srcDir, err := ioutil.TempDir("", "")
			Expect(err).ToNot(HaveOccurred())
			destDir, err := ioutil.TempDir("", "")
			Expect(err).ToNot(HaveOccurred())
			err = ioutil.WriteFile(filepath.Join(srcDir, "tmpFile1"), []byte("Hello, Binky"), 0600)
			Expect(err).ToNot(HaveOccurred())
			err = ioutil.WriteFile(filepath.Join(srcDir, "tmpFile2"), []byte("Bananarama"), 0600)
			Expect(err).ToNot(HaveOccurred())

			zip1 = filepath.Join(destDir, "hello")
			err = zipit(srcDir, zip1, "Hello World Prefix")
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			err := os.RemoveAll(srcDir)
			Expect(err).ToNot(HaveOccurred())
			err = os.RemoveAll(destDir)
			Expect(err).ToNot(HaveOccurred())
		})

		// test create reader, confirm file 1 = hello world, 2 = banana
		It("can do stuff", func() {
			// make a 'Data' for zip1 and zip2
			file1, err := os.Open(zip1)
			Expect(err).ToNot(HaveOccurred())
			defer file1.Close()

			stat1, err := file1.Stat()
			Expect(err).ToNot(HaveOccurred())

			// Compare content
			reader1, err := NewReader(file1, stat1.Size())
			Expect(err).ToNot(HaveOccurred())

			Expect(reader1.File).To(HaveLen(3))
			Expect(reader1.File[1].Name).To(Equal("/tmpFile1"))
			Expect(reader1.File[2].Name).To(Equal("/tmpFile2"))

			f, err := reader1.File[1].Open()
			Expect(err).ToNot(HaveOccurred())
			body, err := ioutil.ReadAll(f)
			Expect(err).ToNot(HaveOccurred())
			Expect(f.Close()).ToNot(HaveOccurred())
			Expect(body).To(Equal([]byte("Hello, Binky")))

			f, err = reader1.File[2].Open()
			Expect(err).ToNot(HaveOccurred())
			body, err = ioutil.ReadAll(f)
			Expect(err).ToNot(HaveOccurred())
			Expect(f.Close()).ToNot(HaveOccurred())
			Expect(body).To(Equal([]byte("Bananarama")))
		})
	})

	Describe("FirstLocalFileHeaderSignature", func() {
		It("returns the position of the first local file header signature", func() {
			data := []byte{
				0x00, 0x40, 0x12, 0x50,
				0x11, 0x50, 0x4b, 0x22,
				0x00, 0x40, 0x12, 0x50,
				0x11, 0x50, 0x4b, 0x22,
				0x11, 0x22, 0x50, 0x4b,
				0x50, 0x4b, 0x03, 0x04, // first local file header signature
				0x11, 0x50, 0x4b, 0x22,
			}
			loc, err := FirstLocalFileHeaderSignature(bytes.NewReader(data))
			Expect(err).ToNot(HaveOccurred())
			Expect(loc).To(BeNumerically("==", 20))
		})
	})
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

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})

	return err
}
