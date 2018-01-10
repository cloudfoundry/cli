// +build !windows,!linux

package sharedaction_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	. "code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/sharedaction/sharedactionfakes"
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
		fakeConfig = new(sharedactionfakes.FakeConfig)
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

		err = ioutil.WriteFile(filepath.Join(subDir, "tmpFile1"), []byte("why hello"), 0644)
		Expect(err).ToNot(HaveOccurred())

		err = ioutil.WriteFile(filepath.Join(srcDir, "tmpFile2"), []byte("Hello, Binky"), 0751)
		Expect(err).ToNot(HaveOccurred())

		err = ioutil.WriteFile(filepath.Join(srcDir, "tmpFile3"), []byte("Bananarama"), 0655)
		Expect(err).ToNot(HaveOccurred())

		err = os.Symlink("file-that-may-or-may-not-exist", filepath.Join(srcDir, "symlink1"))
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		Expect(os.RemoveAll(srcDir)).ToNot(HaveOccurred())
	})

	Describe("GatherArchiveResources", func() {
		var (
			archive string

			resources  []Resource
			executeErr error
		)

		BeforeEach(func() {
			tmpfile, err := ioutil.TempFile("", "example")
			Expect(err).ToNot(HaveOccurred())
			archive = tmpfile.Name()
			Expect(tmpfile.Close()).ToNot(HaveOccurred())
		})

		JustBeforeEach(func() {
			err := zipit(srcDir, archive, "")
			Expect(err).ToNot(HaveOccurred())

			resources, executeErr = actor.GatherArchiveResources(archive)
		})

		AfterEach(func() {
			Expect(os.RemoveAll(archive)).ToNot(HaveOccurred())
		})

		Context("when there is a symlinked file in the archive", func() {
			It("gathers a list of all files in a source archive", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(resources).To(Equal(
					[]Resource{
						{Filename: "/", Mode: DefaultFolderPermissions},
						{Filename: "/level1/", Mode: DefaultFolderPermissions},
						{Filename: "/level1/level2/", Mode: DefaultFolderPermissions},
						{Filename: "/level1/level2/tmpFile1", SHA1: "9e36efec86d571de3a38389ea799a796fe4782f4", Size: 9, Mode: DefaultArchiveFilePermissions},
						{Filename: "/symlink1", Mode: 0755 | os.ModeSymlink},
						{Filename: "/tmpFile2", SHA1: "e594bdc795bb293a0e55724137e53a36dc0d9e95", Size: 12, Mode: DefaultArchiveFilePermissions},
						{Filename: "/tmpFile3", SHA1: "f4c9ca85f3e084ffad3abbdabbd2a890c034c879", Size: 10, Mode: DefaultArchiveFilePermissions},
					}))
			})
		})
	})

	Describe("GatherDirectoryResources", func() {
		var (
			gatheredResources []Resource
			executeErr        error
		)

		JustBeforeEach(func() {
			gatheredResources, executeErr = actor.GatherDirectoryResources(srcDir)
		})

		Context("when a symlink file points to an existing file", func() {
			BeforeEach(func() {
				err := ioutil.WriteFile(filepath.Join(srcDir, "file-that-may-or-may-not-exist"), []byte("Bananarama"), 0655)
				Expect(err).ToNot(HaveOccurred())
			})

			It("does not open the symlink but gathers the name and mode", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(gatheredResources).To(Equal(
					[]Resource{
						{Filename: "file-that-may-or-may-not-exist", SHA1: "f4c9ca85f3e084ffad3abbdabbd2a890c034c879", Size: 10, Mode: 0655},
						{Filename: "level1", Mode: DefaultFolderPermissions},
						{Filename: "level1/level2", Mode: DefaultFolderPermissions},
						{Filename: "level1/level2/tmpFile1", SHA1: "9e36efec86d571de3a38389ea799a796fe4782f4", Size: 9, Mode: 0644},
						{Filename: "symlink1", Mode: 0755 | os.ModeSymlink},
						{Filename: "tmpFile2", SHA1: "e594bdc795bb293a0e55724137e53a36dc0d9e95", Size: 12, Mode: 0751},
						{Filename: "tmpFile3", SHA1: "f4c9ca85f3e084ffad3abbdabbd2a890c034c879", Size: 10, Mode: 0655},
					}))
			})
		})

		Context("when a symlink file points to a file that does not exist", func() {
			It("does not open the symlink but gathers the name and mode", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(gatheredResources).To(Equal(
					[]Resource{
						{Filename: "level1", Mode: DefaultFolderPermissions},
						{Filename: "level1/level2", Mode: DefaultFolderPermissions},
						{Filename: "level1/level2/tmpFile1", SHA1: "9e36efec86d571de3a38389ea799a796fe4782f4", Size: 9, Mode: 0644},
						{Filename: "symlink1", Mode: 0755 | os.ModeSymlink},
						{Filename: "tmpFile2", SHA1: "e594bdc795bb293a0e55724137e53a36dc0d9e95", Size: 12, Mode: 0751},
						{Filename: "tmpFile3", SHA1: "f4c9ca85f3e084ffad3abbdabbd2a890c034c879", Size: 10, Mode: 0655},
					}))
			})
		})
	})
})
