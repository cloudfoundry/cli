package sharedaction_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/sharedaction/sharedactionfakes"
	"code.cloudfoundry.org/ykk"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Resource Actions", func() {
	var (
		actor      *Actor
		fakeConfig *sharedactionfakes.FakeConfig
		srcDir     string
	)

	BeforeEach(func() {
		fakeConfig = &sharedactionfakes.FakeConfig{}
		actor = NewActor(fakeConfig)

		// Creates the following directory structure:
		// level1/level2/tmpFile1
		// tmpfile2
		// tmpfile3

		var err error
		srcDir, err = ioutil.TempDir("", "v2-resource-actions")
		Expect(err).ToNot(HaveOccurred())

		subDir := filepath.Join(srcDir, "level1", "level2")
		err = os.MkdirAll(subDir, 0777)
		Expect(err).ToNot(HaveOccurred())

		err = ioutil.WriteFile(filepath.Join(subDir, "tmpFile1"), []byte("why hello"), 0666)
		Expect(err).ToNot(HaveOccurred())

		err = ioutil.WriteFile(filepath.Join(srcDir, "tmpFile2"), []byte("Hello, Binky"), 0666)
		Expect(err).ToNot(HaveOccurred())

		err = ioutil.WriteFile(filepath.Join(srcDir, "tmpFile3"), []byte("Bananarama"), 0666)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		Expect(os.RemoveAll(srcDir)).ToNot(HaveOccurred())
	})

	Describe("GatherArchiveResources", func() {
		Context("when the archive exists", func() {
			var (
				archive string

				resources  []Resource
				executeErr error
			)

			BeforeEach(func() {
				tmpfile, err := ioutil.TempFile("", "gather-archive-resource-test")
				Expect(err).ToNot(HaveOccurred())
				defer tmpfile.Close()
				archive = tmpfile.Name()
			})

			JustBeforeEach(func() {
				err := zipit(srcDir, archive, "")
				Expect(err).ToNot(HaveOccurred())

				resources, executeErr = actor.GatherArchiveResources(archive)
			})

			AfterEach(func() {
				Expect(os.RemoveAll(archive)).ToNot(HaveOccurred())
			})

			It("gathers a list of all files in a source archive", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(resources).To(Equal(
					[]Resource{
						{Filename: "/", Mode: DefaultFolderPermissions},
						{Filename: "/level1/", Mode: DefaultFolderPermissions},
						{Filename: "/level1/level2/", Mode: DefaultFolderPermissions},
						{Filename: "/level1/level2/tmpFile1", SHA1: "9e36efec86d571de3a38389ea799a796fe4782f4", Size: 9, Mode: DefaultArchiveFilePermissions},
						{Filename: "/tmpFile2", SHA1: "e594bdc795bb293a0e55724137e53a36dc0d9e95", Size: 12, Mode: DefaultArchiveFilePermissions},
						{Filename: "/tmpFile3", SHA1: "f4c9ca85f3e084ffad3abbdabbd2a890c034c879", Size: 10, Mode: DefaultArchiveFilePermissions},
					}))
			})

			Context("when the file is a symlink to an archive", func() {
				var symlinkToArchive string

				BeforeEach(func() {
					tempFile, err := ioutil.TempFile("", "symlink-to-archive")
					Expect(err).ToNot(HaveOccurred())
					Expect(tempFile.Close()).To(Succeed())
					symlinkToArchive = tempFile.Name()
					Expect(os.Remove(symlinkToArchive)).To(Succeed())

					Expect(os.Symlink(archive, symlinkToArchive)).To(Succeed())
				})

				JustBeforeEach(func() {
					resources, executeErr = actor.GatherArchiveResources(symlinkToArchive)
				})

				AfterEach(func() {
					Expect(os.Remove(symlinkToArchive)).To(Succeed())
				})

				It("gathers a list of all files in a source archive", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(resources).To(Equal(
						[]Resource{
							{Filename: "/", Mode: DefaultFolderPermissions},
							{Filename: "/level1/", Mode: DefaultFolderPermissions},
							{Filename: "/level1/level2/", Mode: DefaultFolderPermissions},
							{Filename: "/level1/level2/tmpFile1", SHA1: "9e36efec86d571de3a38389ea799a796fe4782f4", Size: 9, Mode: DefaultArchiveFilePermissions},
							{Filename: "/tmpFile2", SHA1: "e594bdc795bb293a0e55724137e53a36dc0d9e95", Size: 12, Mode: DefaultArchiveFilePermissions},
							{Filename: "/tmpFile3", SHA1: "f4c9ca85f3e084ffad3abbdabbd2a890c034c879", Size: 10, Mode: DefaultArchiveFilePermissions},
						}))
				})
			})

			Context("when a .cfignore file exists in the archive", func() {
				BeforeEach(func() {
					err := ioutil.WriteFile(filepath.Join(srcDir, ".cfignore"), []byte("level2"), 0655)
					Expect(err).ToNot(HaveOccurred())
				})

				It("excludes all patterns of files mentioned in .cfignore", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(resources).To(Equal(
						[]Resource{
							{Filename: "/", Mode: DefaultFolderPermissions},
							{Filename: "/level1/", Mode: DefaultFolderPermissions},
							{Filename: "/tmpFile2", SHA1: "e594bdc795bb293a0e55724137e53a36dc0d9e95", Size: 12, Mode: DefaultArchiveFilePermissions},
							{Filename: "/tmpFile3", SHA1: "f4c9ca85f3e084ffad3abbdabbd2a890c034c879", Size: 10, Mode: DefaultArchiveFilePermissions},
						}))
				})
			})

			Context("when default ignored files exist in the archive", func() {
				BeforeEach(func() {
					for _, filename := range DefaultIgnoreLines {
						if filename != ".cfignore" {
							err := ioutil.WriteFile(filepath.Join(srcDir, filename), nil, 0655)
							Expect(err).ToNot(HaveOccurred())
							err = ioutil.WriteFile(filepath.Join(srcDir, "level1", filename), nil, 0655)
							Expect(err).ToNot(HaveOccurred())
						}
					}
				})

				It("excludes all default files", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(resources).To(Equal(
						[]Resource{
							{Filename: "/", Mode: DefaultFolderPermissions},
							{Filename: "/level1/", Mode: DefaultFolderPermissions},
							{Filename: "/level1/level2/", Mode: DefaultFolderPermissions},
							{Filename: "/level1/level2/tmpFile1", SHA1: "9e36efec86d571de3a38389ea799a796fe4782f4", Size: 9, Mode: DefaultArchiveFilePermissions},
							{Filename: "/tmpFile2", SHA1: "e594bdc795bb293a0e55724137e53a36dc0d9e95", Size: 12, Mode: DefaultArchiveFilePermissions},
							{Filename: "/tmpFile3", SHA1: "f4c9ca85f3e084ffad3abbdabbd2a890c034c879", Size: 10, Mode: DefaultArchiveFilePermissions},
						}))
				})
			})
		})

		Context("when the archive does not exist", func() {
			It("returns an error if the file is problematic", func() {
				_, err := actor.GatherArchiveResources("/does/not/exist")
				Expect(os.IsNotExist(err)).To(BeTrue())
			})
		})
	})

	Describe("GatherDirectoryResources", func() {
		Context("when files exist in the directory", func() {
			var (
				gatheredResources []Resource
				executeErr        error
			)

			JustBeforeEach(func() {
				gatheredResources, executeErr = actor.GatherDirectoryResources(srcDir)
			})

			It("gathers a list of all directories files in a source directory", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(gatheredResources).To(Equal(
					[]Resource{
						{Filename: "level1", Mode: DefaultFolderPermissions},
						{Filename: "level1/level2", Mode: DefaultFolderPermissions},
						{Filename: "level1/level2/tmpFile1", SHA1: "9e36efec86d571de3a38389ea799a796fe4782f4", Size: 9, Mode: 0766},
						{Filename: "tmpFile2", SHA1: "e594bdc795bb293a0e55724137e53a36dc0d9e95", Size: 12, Mode: 0766},
						{Filename: "tmpFile3", SHA1: "f4c9ca85f3e084ffad3abbdabbd2a890c034c879", Size: 10, Mode: 0766},
					}))
			})

			Context("when the provided path is a symlink to the directory", func() {
				var tmpDir string

				BeforeEach(func() {
					tmpDir = srcDir

					tmpFile, err := ioutil.TempFile("", "symlink-file-")
					Expect(err).ToNot(HaveOccurred())
					Expect(tmpFile.Close()).To(Succeed())

					srcDir = tmpFile.Name()
					Expect(os.Remove(srcDir)).To(Succeed())
					Expect(os.Symlink(tmpDir, srcDir)).To(Succeed())
				})

				AfterEach(func() {
					Expect(os.RemoveAll(tmpDir)).To(Succeed())
				})

				It("gathers a list of all directories files in a source directory", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(gatheredResources).To(Equal(
						[]Resource{
							{Filename: "level1", Mode: DefaultFolderPermissions},
							{Filename: "level1/level2", Mode: DefaultFolderPermissions},
							{Filename: "level1/level2/tmpFile1", SHA1: "9e36efec86d571de3a38389ea799a796fe4782f4", Size: 9, Mode: 0766},
							{Filename: "tmpFile2", SHA1: "e594bdc795bb293a0e55724137e53a36dc0d9e95", Size: 12, Mode: 0766},
							{Filename: "tmpFile3", SHA1: "f4c9ca85f3e084ffad3abbdabbd2a890c034c879", Size: 10, Mode: 0766},
						}))
				})
			})

			Context("when a .cfignore file exists in the sourceDir", func() {
				Context("with relative paths", func() {
					BeforeEach(func() {
						err := ioutil.WriteFile(filepath.Join(srcDir, ".cfignore"), []byte("level2"), 0666)
						Expect(err).ToNot(HaveOccurred())
					})

					It("excludes all patterns of files mentioned in .cfignore", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(gatheredResources).To(Equal(
							[]Resource{
								{Filename: "level1", Mode: DefaultFolderPermissions},
								{Filename: "tmpFile2", SHA1: "e594bdc795bb293a0e55724137e53a36dc0d9e95", Size: 12, Mode: 0766},
								{Filename: "tmpFile3", SHA1: "f4c9ca85f3e084ffad3abbdabbd2a890c034c879", Size: 10, Mode: 0766},
							}))
					})
				})

				Context("with absolute paths - where '/' == sourceDir", func() {
					BeforeEach(func() {
						err := ioutil.WriteFile(filepath.Join(srcDir, ".cfignore"), []byte("/level1/level2"), 0666)
						Expect(err).ToNot(HaveOccurred())
					})

					It("excludes all patterns of files mentioned in .cfignore", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(gatheredResources).To(Equal(
							[]Resource{
								{Filename: "level1", Mode: DefaultFolderPermissions},
								{Filename: "tmpFile2", SHA1: "e594bdc795bb293a0e55724137e53a36dc0d9e95", Size: 12, Mode: 0766},
								{Filename: "tmpFile3", SHA1: "f4c9ca85f3e084ffad3abbdabbd2a890c034c879", Size: 10, Mode: 0766},
							}))
					})
				})
			})

			Context("when default ignored files exist in the app dir", func() {
				BeforeEach(func() {
					for _, filename := range DefaultIgnoreLines {
						if filename != ".cfignore" {
							err := ioutil.WriteFile(filepath.Join(srcDir, filename), nil, 0655)
							Expect(err).ToNot(HaveOccurred())
							err = ioutil.WriteFile(filepath.Join(srcDir, "level1", filename), nil, 0655)
							Expect(err).ToNot(HaveOccurred())
						}
					}
				})

				It("excludes all default files", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(gatheredResources).To(Equal(
						[]Resource{
							{Filename: "level1", Mode: DefaultFolderPermissions},
							{Filename: "level1/level2", Mode: DefaultFolderPermissions},
							{Filename: "level1/level2/tmpFile1", SHA1: "9e36efec86d571de3a38389ea799a796fe4782f4", Size: 9, Mode: 0766},
							{Filename: "tmpFile2", SHA1: "e594bdc795bb293a0e55724137e53a36dc0d9e95", Size: 12, Mode: 0766},
							{Filename: "tmpFile3", SHA1: "f4c9ca85f3e084ffad3abbdabbd2a890c034c879", Size: 10, Mode: 0766},
						}))
				})
			})

			Context("when trace files are in the source directory", func() {
				BeforeEach(func() {
					traceFilePath := filepath.Join(srcDir, "i-am-trace.txt")
					err := ioutil.WriteFile(traceFilePath, nil, 0666)
					Expect(err).ToNot(HaveOccurred())

					fakeConfig.VerboseReturns(false, []string{traceFilePath, "C:\\some-other-path"})
				})

				It("excludes all of the trace files", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(gatheredResources).To(Equal(
						[]Resource{
							{Filename: "level1", Mode: DefaultFolderPermissions},
							{Filename: "level1/level2", Mode: DefaultFolderPermissions},
							{Filename: "level1/level2/tmpFile1", SHA1: "9e36efec86d571de3a38389ea799a796fe4782f4", Size: 9, Mode: 0766},
							{Filename: "tmpFile2", SHA1: "e594bdc795bb293a0e55724137e53a36dc0d9e95", Size: 12, Mode: 0766},
							{Filename: "tmpFile3", SHA1: "f4c9ca85f3e084ffad3abbdabbd2a890c034c879", Size: 10, Mode: 0766},
						}))
				})
			})
		})

		Context("when the directory is empty", func() {
			var emptyDir string

			BeforeEach(func() {
				var err error
				emptyDir, err = ioutil.TempDir("", "v2-resource-actions-empty")
				Expect(err).ToNot(HaveOccurred())
			})

			AfterEach(func() {
				Expect(os.RemoveAll(emptyDir)).ToNot(HaveOccurred())
			})

			It("returns an EmptyDirectoryError", func() {
				_, err := actor.GatherDirectoryResources(emptyDir)
				Expect(err).To(MatchError(actionerror.EmptyDirectoryError{Path: emptyDir}))
			})
		})
	})

	Describe("ZipDirectoryResources", func() {
		var (
			resultZip  string
			resources  []Resource
			executeErr error
		)

		BeforeEach(func() {
			resources = []Resource{
				{Filename: "level1", Mode: DefaultFolderPermissions},
				{Filename: "level1/level2", Mode: DefaultFolderPermissions},
				{Filename: "level1/level2/tmpFile1", SHA1: "9e36efec86d571de3a38389ea799a796fe4782f4", Size: 9, Mode: 0766},
				{Filename: "tmpFile2", SHA1: "e594bdc795bb293a0e55724137e53a36dc0d9e95", Size: 12, Mode: 0766},
				{Filename: "tmpFile3", SHA1: "f4c9ca85f3e084ffad3abbdabbd2a890c034c879", Size: 10, Mode: 0766},
			}
		})

		JustBeforeEach(func() {
			resultZip, executeErr = actor.ZipDirectoryResources(srcDir, resources)
		})

		AfterEach(func() {
			err := os.RemoveAll(srcDir)
			Expect(err).ToNot(HaveOccurred())

			err = os.RemoveAll(resultZip)
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when zipping on windows", func() {
			It("zips the directory and sets all the file modes to 07XX", func() {
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
				Expect(reader.File[2].Mode()).To(Equal(os.FileMode(0766)), reader.File[2].Name)
				Expect(reader.File[3].Mode()).To(Equal(os.FileMode(0766)), reader.File[3].Name)
				Expect(reader.File[4].Mode()).To(Equal(os.FileMode(0766)), reader.File[4].Name)
			})
		})
	})
})
