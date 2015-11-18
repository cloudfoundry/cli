package app_files_test

import (
	"archive/zip"
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	. "github.com/cloudfoundry/cli/cf/app_files"
	"github.com/cloudfoundry/gofileutils/fileutils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func readFile(file *os.File) []byte {
	bytes, err := ioutil.ReadAll(file)
	Expect(err).NotTo(HaveOccurred())
	return bytes
}

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

var _ = Describe("Zipper", func() {
	var filesInZip = []string{
		"foo.txt",
		"fooDir/",
		"fooDir/bar/",
		"lastDir/",
		"subDir/",
		"subDir/bar.txt",
		"subDir/otherDir/",
		"subDir/otherDir/file.txt",
	}

	It("zips directories", func() {
		fileutils.TempFile("zip_test", func(zipFile *os.File, err error) {
			workingDir, err := os.Getwd()
			Expect(err).NotTo(HaveOccurred())

			dir := filepath.Join(workingDir, "../../fixtures/zip/")
			err = os.Chmod(filepath.Join(dir, "subDir/bar.txt"), 0666)
			Expect(err).NotTo(HaveOccurred())

			zipper := ApplicationZipper{}
			err = zipper.Zip(dir, zipFile)
			Expect(err).NotTo(HaveOccurred())

			fileStat, err := zipFile.Stat()
			Expect(err).NotTo(HaveOccurred())

			reader, err := zip.NewReader(zipFile, fileStat.Size())
			Expect(err).NotTo(HaveOccurred())

			filenames := []string{}
			for _, file := range reader.File {
				filenames = append(filenames, file.Name)
			}

			Expect(filenames).To(Equal(filesInZip))

			readFileInZip := func(index int) (string, string) {
				buf := &bytes.Buffer{}
				file := reader.File[index]
				fReader, err := file.Open()
				_, err = io.Copy(buf, fReader)

				Expect(err).NotTo(HaveOccurred())

				return file.Name, string(buf.Bytes())
			}

			Expect(err).NotTo(HaveOccurred())

			name, contents := readFileInZip(0)
			Expect(name).To(Equal("foo.txt"))
			Expect(contents).To(Equal("This is a simple text file."))

			name, contents = readFileInZip(5)
			Expect(name).To(Equal("subDir/bar.txt"))
			Expect(contents).To(Equal("I am in a subdirectory."))
			Expect(reader.File[5].FileInfo().Mode()).To(Equal(os.FileMode(0666)))
		})
	})

	It("is a no-op for a zipfile", func() {
		fileutils.TempFile("zip_test", func(zipFile *os.File, err error) {
			dir, err := os.Getwd()
			Expect(err).NotTo(HaveOccurred())

			zipper := ApplicationZipper{}
			fixture := filepath.Join(dir, "../../fixtures/applications/example-app.zip")
			err = zipper.Zip(fixture, zipFile)
			Expect(err).NotTo(HaveOccurred())

			zippedFile, err := os.Open(fixture)
			Expect(err).NotTo(HaveOccurred())
			Expect(readFile(zipFile)).To(Equal(readFile(zippedFile)))
		})
	})

	It("returns an error when zipping fails", func() {
		fileutils.TempFile("zip_test", func(zipFile *os.File, err error) {
			zipper := ApplicationZipper{}
			err = zipper.Zip("/a/bogus/directory", zipFile)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("open /a/bogus/directory"))
		})
	})

	It("returns an error when the directory is empty", func() {
		fileutils.TempFile("zip_test", func(zipFile *os.File, err error) {
			fileutils.TempDir("zip_test", func(emptyDir string, err error) {
				zipper := ApplicationZipper{}
				err = zipper.Zip(emptyDir, zipFile)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("is empty"))
			})
		})
	})

	Describe("IsZipFile", func() {
		var (
			inDir, outDir string
			zipper        ApplicationZipper
		)

		AfterEach(func() {
			os.RemoveAll(inDir)
			os.RemoveAll(outDir)
		})

		Context("when given a zip without prefix bytes", func() {
			BeforeEach(func() {
				var err error
				inDir, err = ioutil.TempDir("", "zipper-unzip-in")
				Expect(err).NotTo(HaveOccurred())

				err = ioutil.WriteFile(path.Join(inDir, "file1"), []byte("file-1-contents"), 0664)
				Expect(err).NotTo(HaveOccurred())

				outDir, err = ioutil.TempDir("", "zipper-unzip-out")
				Expect(err).NotTo(HaveOccurred())

				err = zipit(path.Join(inDir, "/"), path.Join(outDir, "out.zip"), "")
				Expect(err).NotTo(HaveOccurred())

				zipper = ApplicationZipper{}
			})

			It("returns true", func() {
				Expect(zipper.IsZipFile(path.Join(outDir, "out.zip"))).To(BeTrue())
			})
		})

		Context("when given a zip with prefix bytes", func() {
			BeforeEach(func() {
				var err error
				inDir, err = ioutil.TempDir("", "zipper-unzip-in")
				Expect(err).NotTo(HaveOccurred())

				err = ioutil.WriteFile(path.Join(inDir, "file1"), []byte("file-1-contents"), 0664)
				Expect(err).NotTo(HaveOccurred())

				outDir, err = ioutil.TempDir("", "zipper-unzip-out")
				Expect(err).NotTo(HaveOccurred())

				err = zipit(path.Join(inDir, "/"), path.Join(outDir, "out.zip"), "prefix-bytes")
				Expect(err).NotTo(HaveOccurred())

				zipper = ApplicationZipper{}
			})

			It("returns true", func() {
				Expect(zipper.IsZipFile(path.Join(outDir, "out.zip"))).To(BeTrue())
			})
		})

		Context("when given a file that is not a zip", func() {
			var fileName string

			BeforeEach(func() {
				f, err := ioutil.TempFile("", "zipper-test")
				Expect(err).NotTo(HaveOccurred())

				fi, err := f.Stat()
				Expect(err).NotTo(HaveOccurred())
				fileName = fi.Name()
			})

			AfterEach(func() {
				defer os.RemoveAll(fileName)
			})

			It("returns false", func() {
				Expect(zipper.IsZipFile(fileName)).To(BeFalse())
			})
		})

		Context("when given a directory", func() {
			var dirName string

			BeforeEach(func() {
				var err error
				dirName, err = ioutil.TempDir("", "zipper-test")
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				defer os.RemoveAll(dirName)
			})

			It("returns false", func() {
				Expect(zipper.IsZipFile(dirName)).To(BeFalse())
			})
		})
	})

	Describe(".Unzip", func() {
		var (
			inDir, outDir string
			zipper        ApplicationZipper
		)

		AfterEach(func() {
			os.RemoveAll(inDir)
			os.RemoveAll(outDir)
		})

		Context("when the zipfile has prefix bytes", func() {
			BeforeEach(func() {
				var err error
				inDir, err = ioutil.TempDir("", "zipper-unzip-in")
				Expect(err).NotTo(HaveOccurred())

				err = ioutil.WriteFile(path.Join(inDir, "file1"), []byte("file-1-contents"), 0664)
				Expect(err).NotTo(HaveOccurred())

				outDir, err = ioutil.TempDir("", "zipper-unzip-out")
				Expect(err).NotTo(HaveOccurred())

				err = zipit(path.Join(inDir, "/"), path.Join(outDir, "out.zip"), "prefix-bytes")
				Expect(err).NotTo(HaveOccurred())

				zipper = ApplicationZipper{}
			})

			It("successfully extracts the zip", func() {
				destDir, err := ioutil.TempDir("", "dest-dir")
				Expect(err).NotTo(HaveOccurred())

				defer os.RemoveAll(destDir)

				err = zipper.Unzip(path.Join(outDir, "out.zip"), destDir)
				Expect(err).NotTo(HaveOccurred())

				_, err = os.Stat(filepath.Join(destDir, "file1"))
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the zipfile has an empty directory", func() {
			BeforeEach(func() {
				var err error
				inDir, err = ioutil.TempDir("", "zipper-unzip-in")
				Expect(err).NotTo(HaveOccurred())

				err = ioutil.WriteFile(path.Join(inDir, "file1"), []byte("file-1-contents"), 0664)
				Expect(err).NotTo(HaveOccurred())

				err = os.MkdirAll(path.Join(inDir, "dir1"), os.ModeDir|os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				err = ioutil.WriteFile(path.Join(inDir, "dir1", "file2"), []byte("file-2-contents"), 0644)
				Expect(err).NotTo(HaveOccurred())

				err = os.MkdirAll(path.Join(inDir, "dir2"), os.ModeDir|os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				outDir, err = ioutil.TempDir("", "zipper-unzip-out")
				Expect(err).NotTo(HaveOccurred())

				err = zipit(path.Join(inDir, "/"), path.Join(outDir, "out.zip"), "")
				Expect(err).NotTo(HaveOccurred())

				zipper = ApplicationZipper{}
			})

			It("includes all entries from the zip file in the destination", func() {
				destDir, err := ioutil.TempDir("", "dest-dir")
				Expect(err).NotTo(HaveOccurred())

				defer os.RemoveAll(destDir)

				err = zipper.Unzip(path.Join(outDir, "out.zip"), destDir)
				Expect(err).NotTo(HaveOccurred())

				expected := []string{
					"file1",
					"dir1/",
					"dir1/file2",
					"dir2",
				}

				for _, f := range expected {
					_, err := os.Stat(filepath.Join(destDir, f))
					Expect(err).NotTo(HaveOccurred())
				}
			})
		})
	})

	Describe(".GetZipSize", func() {
		var zipper = ApplicationZipper{}

		It("returns the size of the zip file", func() {
			dir, err := os.Getwd()
			Expect(err).NotTo(HaveOccurred())
			zipFile := filepath.Join(dir, "../../fixtures/applications/example-app.zip")

			file, err := os.Open(zipFile)
			Expect(err).NotTo(HaveOccurred())

			fileSize, err := zipper.GetZipSize(file)
			Expect(err).NotTo(HaveOccurred())
			Expect(fileSize).To(Equal(int64(1803)))
		})

		It("returns  an error if the zip file cannot be found", func() {
			tmpFile, _ := os.Open("fooBar")
			_, sizeErr := zipper.GetZipSize(tmpFile)
			Expect(sizeErr).To(HaveOccurred())
		})
	})

})
