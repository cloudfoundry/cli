package sharedaction_test

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/sharedaction/sharedactionfakes"
	"code.cloudfoundry.org/ykk"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Resource Actions", func() {
	var (
		fakeConfig *sharedactionfakes.FakeConfig
		actor      *Actor
		srcDir     string
	)

	BeforeEach(func() {
		fakeConfig = new(sharedactionfakes.FakeConfig)
		actor = NewActor(fakeConfig)

		// Creates the following directory structure:
		// level1/level2/tmpFile1
		// tmpfile2
		// tmpfile3

		var err error
		srcDir, err = os.MkdirTemp("", "resource-actions-test")
		Expect(err).ToNot(HaveOccurred())

		subDir := filepath.Join(srcDir, "level1", "level2")
		err = os.MkdirAll(subDir, 0777)
		Expect(err).ToNot(HaveOccurred())

		err = os.WriteFile(filepath.Join(subDir, "tmpFile1"), []byte("why hello"), 0600)
		Expect(err).ToNot(HaveOccurred())

		err = os.WriteFile(filepath.Join(srcDir, "tmpFile2"), []byte("Hello, Binky"), 0600)
		Expect(err).ToNot(HaveOccurred())

		err = os.WriteFile(filepath.Join(srcDir, "tmpFile3"), []byte("Bananarama"), 0600)
		Expect(err).ToNot(HaveOccurred())

		relativePath, err := filepath.Rel(srcDir, subDir)
		Expect(err).ToNot(HaveOccurred())

		// ./symlink -> ./level1/level2/tmpfile1
		err = os.Symlink(filepath.Join(relativePath, "tmpFile1"), filepath.Join(srcDir, "symlink1"))
		Expect(err).ToNot(HaveOccurred())

		// ./level1/level2/symlink2 -> ../../tmpfile2
		err = os.Symlink("../../tmpfile2", filepath.Join(subDir, "symlink2"))
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		Expect(os.RemoveAll(srcDir)).ToNot(HaveOccurred())
	})

	Describe("SharedToV3Resource", func() {

		var returnedV3Resource V3Resource
		var sharedResource = Resource{Filename: "file1", SHA1: "a43rknl", Mode: os.FileMode(0644), Size: 100000}

		JustBeforeEach(func() {
			returnedV3Resource = sharedResource.ToV3Resource()
		})

		It("returns a ccv3 Resource", func() {
			Expect(returnedV3Resource).To(Equal(V3Resource{
				FilePath: "file1",
				Checksum: ccv3.Checksum{
					Value: "a43rknl",
				},
				SizeInBytes: 100000,
				Mode:        os.FileMode(0644),
			}))
		})

	})

	Describe("ToSharedResource", func() {

		var returnedSharedResource Resource
		var v3Resource = V3Resource{
			FilePath: "file1",
			Checksum: ccv3.Checksum{
				Value: "a43rknl",
			},
			SizeInBytes: 100000,
			Mode:        os.FileMode(0644),
		}

		JustBeforeEach(func() {
			returnedSharedResource = v3Resource.ToV2Resource()
		})

		It("returns a ccv3 Resource", func() {
			Expect(returnedSharedResource).To(Equal(Resource{Filename: "file1", SHA1: "a43rknl", Mode: os.FileMode(0644), Size: 100000}))

		})
	})

	Describe("GatherArchiveResources", func() {
		// tests are under resource_unix_test.go and resource_windows_test.go
	})

	Describe("GatherDirectoryResources", func() {
		// tests are under resource_unix_test.go and resource_windows_test.go
	})

	Describe("ReadArchive", func() {
		var (
			archivePath string
			executeErr  error

			readCloser io.ReadCloser
			fileSize   int64
		)

		JustBeforeEach(func() {
			readCloser, fileSize, executeErr = actor.ReadArchive(archivePath)
		})

		AfterEach(func() {
			Expect(os.RemoveAll(archivePath)).ToNot(HaveOccurred())
		})

		When("the archive can be accessed properly", func() {
			BeforeEach(func() {
				tmpfile, err := os.CreateTemp("", "fake-archive")
				Expect(err).ToNot(HaveOccurred())
				_, err = tmpfile.Write([]byte("123456"))
				Expect(err).ToNot(HaveOccurred())
				Expect(tmpfile.Close()).ToNot(HaveOccurred())

				archivePath = tmpfile.Name()
			})

			It("returns zero errors", func() {
				defer readCloser.Close()

				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fileSize).To(BeNumerically("==", 6))

				b := make([]byte, 100)
				size, err := readCloser.Read(b)
				Expect(err).ToNot(HaveOccurred())
				Expect(b[:size]).To(Equal([]byte("123456")))
			})
		})

		When("the archive returns any access errors", func() {
			It("returns the error", func() {
				_, ok := executeErr.(*os.PathError)
				Expect(ok).To(BeTrue())
			})
		})
	})

	Describe("ZipArchiveResources", func() {
		var (
			archive    string
			resultZip  string
			resources  []Resource
			executeErr error
		)

		BeforeEach(func() {
			tmpfile, err := os.CreateTemp("", "zip-archive-resources")
			Expect(err).ToNot(HaveOccurred())
			defer tmpfile.Close()
			archive = tmpfile.Name()

			err = zipit(srcDir, archive, "")
			Expect(err).ToNot(HaveOccurred())
		})

		JustBeforeEach(func() {
			resultZip, executeErr = actor.ZipArchiveResources(archive, resources)
		})

		AfterEach(func() {
			Expect(os.RemoveAll(archive)).ToNot(HaveOccurred())
			Expect(os.RemoveAll(resultZip)).ToNot(HaveOccurred())
		})

		When("the files have not been changed since scanning them", func() {
			When("there are no symlinks", func() {
				BeforeEach(func() {
					resources = []Resource{
						{Filename: "/"},
						{Filename: "/level1/"},
						{Filename: "/level1/level2/"},
						{Filename: "/level1/level2/tmpFile1", SHA1: "9e36efec86d571de3a38389ea799a796fe4782f4"},
						{Filename: "/tmpFile2", SHA1: "e594bdc795bb293a0e55724137e53a36dc0d9e95"},
						// Explicitly skipping /tmpFile3
					}
				})

				It("zips the file and returns a populated resources list", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(resultZip).ToNot(BeEmpty())
					zipFile, err := os.Open(resultZip)
					Expect(err).ToNot(HaveOccurred())
					defer zipFile.Close()

					zipInfo, err := zipFile.Stat()
					Expect(err).ToNot(HaveOccurred())

					reader, err := ykk.NewReader(zipFile, zipInfo.Size())
					Expect(err).ToNot(HaveOccurred())

					Expect(reader.File).To(HaveLen(5))

					Expect(reader.File[0].Name).To(Equal("/"))

					Expect(reader.File[1].Name).To(Equal("/level1/"))

					Expect(reader.File[2].Name).To(Equal("/level1/level2/"))

					Expect(reader.File[3].Name).To(Equal("/level1/level2/tmpFile1"))
					Expect(reader.File[3].Method).To(Equal(zip.Deflate))
					expectFileContentsToEqual(reader.File[3], "why hello")

					Expect(reader.File[4].Name).To(Equal("/tmpFile2"))
					Expect(reader.File[4].Method).To(Equal(zip.Deflate))
					expectFileContentsToEqual(reader.File[4], "Hello, Binky")
				})
			})

			When("there are relative symlink files", func() {
				BeforeEach(func() {
					resources = []Resource{
						{Filename: "/"},
						{Filename: "/level1/"},
						{Filename: "/level1/level2/"},
						{Filename: "/level1/level2/tmpFile1", SHA1: "9e36efec86d571de3a38389ea799a796fe4782f4"},
						{Filename: "/symlink1", Mode: os.ModeSymlink | 0777},
						{Filename: "/level1/level2/symlink2", Mode: os.ModeSymlink | 0777},
					}
				})

				It("zips the file and returns a populated resources list", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(resultZip).ToNot(BeEmpty())
					zipFile, err := os.Open(resultZip)
					Expect(err).ToNot(HaveOccurred())
					defer zipFile.Close()

					zipInfo, err := zipFile.Stat()
					Expect(err).ToNot(HaveOccurred())

					reader, err := ykk.NewReader(zipFile, zipInfo.Size())
					Expect(err).ToNot(HaveOccurred())

					Expect(reader.File).To(HaveLen(6))
					Expect(reader.File[0].Name).To(Equal("/"))
					Expect(reader.File[1].Name).To(Equal("/level1/"))
					Expect(reader.File[2].Name).To(Equal("/level1/level2/"))
					Expect(reader.File[3].Name).To(Equal("/level1/level2/symlink2"))
					Expect(reader.File[4].Name).To(Equal("/level1/level2/tmpFile1"))
					Expect(reader.File[5].Name).To(Equal("/symlink1"))

					expectFileContentsToEqual(reader.File[4], "why hello")
					Expect(reader.File[5].Mode() & os.ModeSymlink).To(Equal(os.ModeSymlink))
					expectFileContentsToEqual(reader.File[5], filepath.FromSlash("level1/level2/tmpFile1"))

					Expect(reader.File[3].Mode() & os.ModeSymlink).To(Equal(os.ModeSymlink))
					expectFileContentsToEqual(reader.File[3], filepath.FromSlash("../../tmpfile2"))
				})
			})
		})

		When("the files have changed since the scanning", func() {
			BeforeEach(func() {
				resources = []Resource{
					{Filename: "/"},
					{Filename: "/level1/"},
					{Filename: "/level1/level2/"},
					{Filename: "/level1/level2/tmpFile1", SHA1: "9e36efec86d571de3a38389ea799a796fe4782f4"},
					{Filename: "/tmpFile2", SHA1: "e594bdc795bb293a0e55724137e53a36dc0d9e95"},
					{Filename: "/tmpFile3", SHA1: "i dunno, 7?"},
				}
			})

			It("returns an FileChangedError", func() {
				Expect(executeErr).To(Equal(actionerror.FileChangedError{Filename: "/tmpFile3"}))
			})
		})
	})

	Describe("ZipDirectoryResources", func() {
		var (
			resultZip  string
			resources  []Resource
			executeErr error
		)

		JustBeforeEach(func() {
			resultZip, executeErr = actor.ZipDirectoryResources(srcDir, resources)
		})

		AfterEach(func() {
			Expect(os.RemoveAll(resultZip)).ToNot(HaveOccurred())
		})

		When("the files have not been changed since scanning them", func() {
			When("there are no symlinks", func() {
				BeforeEach(func() {
					resources = []Resource{
						{Filename: "level1"},
						{Filename: "level1/level2"},
						{Filename: "level1/level2/tmpFile1", SHA1: "9e36efec86d571de3a38389ea799a796fe4782f4"},
						{Filename: "tmpFile2", SHA1: "e594bdc795bb293a0e55724137e53a36dc0d9e95"},
						{Filename: "tmpFile3", SHA1: "f4c9ca85f3e084ffad3abbdabbd2a890c034c879"},
					}
				})

				It("zips the file and returns a populated resources list", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(resultZip).ToNot(BeEmpty())
					zipFile, err := os.Open(resultZip)
					Expect(err).ToNot(HaveOccurred())
					defer zipFile.Close()

					zipInfo, err := zipFile.Stat()
					Expect(err).ToNot(HaveOccurred())

					reader, err := ykk.NewReader(zipFile, zipInfo.Size())
					Expect(err).ToNot(HaveOccurred())

					Expect(reader.File).To(HaveLen(5))
					Expect(reader.File[0].Name).To(Equal("level1/"))

					Expect(reader.File[1].Name).To(Equal("level1/level2/"))

					Expect(reader.File[2].Name).To(Equal("level1/level2/tmpFile1"))
					Expect(reader.File[2].Method).To(Equal(zip.Deflate))
					expectFileContentsToEqual(reader.File[2], "why hello")

					Expect(reader.File[3].Name).To(Equal("tmpFile2"))
					Expect(reader.File[3].Method).To(Equal(zip.Deflate))
					expectFileContentsToEqual(reader.File[3], "Hello, Binky")

					Expect(reader.File[4].Name).To(Equal("tmpFile3"))
					Expect(reader.File[4].Method).To(Equal(zip.Deflate))
					expectFileContentsToEqual(reader.File[4], "Bananarama")
				})
			})

			When("there are relative symlink files", func() {
				BeforeEach(func() {
					resources = []Resource{
						{Filename: "level1"},
						{Filename: "level1/level2"},
						{Filename: "level1/level2/tmpFile1", SHA1: "9e36efec86d571de3a38389ea799a796fe4782f4"},
						{Filename: "symlink1", Mode: os.ModeSymlink},
						{Filename: "level1/level2/symlink2", Mode: os.ModeSymlink},
					}
				})

				It("zips the file and returns a populated resources list", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(resultZip).ToNot(BeEmpty())
					zipFile, err := os.Open(resultZip)
					Expect(err).ToNot(HaveOccurred())
					defer zipFile.Close()

					zipInfo, err := zipFile.Stat()
					Expect(err).ToNot(HaveOccurred())

					reader, err := ykk.NewReader(zipFile, zipInfo.Size())
					Expect(err).ToNot(HaveOccurred())

					Expect(reader.File).To(HaveLen(5))

					Expect(reader.File[0].Name).To(Equal("level1/"))

					Expect(reader.File[1].Name).To(Equal("level1/level2/"))

					Expect(reader.File[2].Name).To(Equal("level1/level2/tmpFile1"))
					expectFileContentsToEqual(reader.File[2], "why hello")

					Expect(reader.File[3].Name).To(Equal("symlink1"))
					Expect(reader.File[3].Mode() & os.ModeSymlink).To(Equal(os.ModeSymlink))
					expectFileContentsToEqual(reader.File[3], filepath.FromSlash("level1/level2/tmpFile1"))

					Expect(reader.File[4].Name).To(Equal("level1/level2/symlink2"))
					Expect(reader.File[4].Mode() & os.ModeSymlink).To(Equal(os.ModeSymlink))
					expectFileContentsToEqual(reader.File[4], filepath.FromSlash("../../tmpfile2"))
				})
			})
		})

		When("the files have changed since the scanning", func() {
			BeforeEach(func() {
				resources = []Resource{
					{Filename: "level1"},
					{Filename: "level1/level2"},
					{Filename: "level1/level2/tmpFile1", SHA1: "9e36efec86d571de3a38389ea799a796fe4782f4"},
					{Filename: "tmpFile2", SHA1: "e594bdc795bb293a0e55724137e53a36dc0d9e95"},
					{Filename: "tmpFile3", SHA1: "i dunno, 7?"},
				}
			})

			It("returns an FileChangedError", func() {
				Expect(executeErr).To(Equal(actionerror.FileChangedError{Filename: filepath.Join(srcDir, "tmpFile3")}))
			})
		})
	})
})

func expectFileContentsToEqual(file *zip.File, expectedContents string) {
	reader, err := file.Open()
	Expect(err).ToNot(HaveOccurred())
	defer reader.Close()

	body, err := io.ReadAll(reader)
	Expect(err).ToNot(HaveOccurred())

	Expect(string(body)).To(Equal(expectedContents))
}
