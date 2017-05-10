// +build !windows

package v2action_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/ykk"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Resource Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
		srcDir                    string
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v2actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil)

		var err error
		srcDir, err = ioutil.TempDir("", "")
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
	})

	Describe("GatherResources", func() {
		It("gathers a list of all directories files in a source directory", func() {
			resources, err := actor.GatherResources(srcDir)
			Expect(err).ToNot(HaveOccurred())

			Expect(resources).To(Equal(
				[]Resource{
					{Filename: "level1"},
					{Filename: "level1/level2"},
					{Filename: "level1/level2/tmpFile1", SHA1: "9e36efec86d571de3a38389ea799a796fe4782f4", Size: 9, Mode: 0644},
					{Filename: "tmpFile2", SHA1: "e594bdc795bb293a0e55724137e53a36dc0d9e95", Size: 12, Mode: 0751},
					{Filename: "tmpFile3", SHA1: "f4c9ca85f3e084ffad3abbdabbd2a890c034c879", Size: 10, Mode: 0655},
				}))
		})
	})

	Describe("ZipResources", func() {
		var (
			resultZip  string
			resources  []Resource
			executeErr error
		)

		BeforeEach(func() {
			resources = []Resource{
				{Filename: "level1"},
				{Filename: "level1/level2"},
				{Filename: "level1/level2/tmpFile1"},
				{Filename: "tmpFile2"},
				{Filename: "tmpFile3"},
			}
		})

		JustBeforeEach(func() {
			resultZip, executeErr = actor.ZipResources(srcDir, resources)
		})

		AfterEach(func() {
			err := os.RemoveAll(srcDir)
			Expect(err).ToNot(HaveOccurred())

			err = os.RemoveAll(resultZip)
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when zipping on UNIX", func() {
			It("zips the directory and keeps the file permissions", func() {
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
				Expect(reader.File[2].Mode()).To(Equal(os.FileMode(0644)))
				Expect(reader.File[3].Mode()).To(Equal(os.FileMode(0751)))
				Expect(reader.File[4].Mode()).To(Equal(os.FileMode(0655)))
			})
		})
	})
})
