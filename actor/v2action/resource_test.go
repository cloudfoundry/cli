package v2action_test

import (
	"archive/zip"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
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
		actor = NewActor(fakeCloudControllerClient, nil, nil)

		var err error
		srcDir, err = ioutil.TempDir("", "resource-actions-test")
		Expect(err).ToNot(HaveOccurred())

		subDir := filepath.Join(srcDir, "level1", "level2")
		err = os.MkdirAll(subDir, 0777)
		Expect(err).ToNot(HaveOccurred())

		err = ioutil.WriteFile(filepath.Join(subDir, "tmpFile1"), []byte("why hello"), 0600)
		Expect(err).ToNot(HaveOccurred())

		err = ioutil.WriteFile(filepath.Join(srcDir, "tmpFile2"), []byte("Hello, Binky"), 0600)
		Expect(err).ToNot(HaveOccurred())

		err = ioutil.WriteFile(filepath.Join(srcDir, "tmpFile3"), []byte("Bananarama"), 0600)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		Expect(os.RemoveAll(srcDir)).ToNot(HaveOccurred())
	})

	Describe("GatherArchiveResources", func() {
		// tests are under resource_unix_test.go and resource_windows_test.go
	})

	Describe("GatherDirectoryResources", func() {
		// tests are under resource_unix_test.go and resource_windows_test.go
	})

	Describe("ResourceMatch", func() {
		var (
			allResources []Resource

			matchedResources   []Resource
			unmatchedResources []Resource
			warnings           Warnings
			executeErr         error
		)

		JustBeforeEach(func() {
			matchedResources, unmatchedResources, warnings, executeErr = actor.ResourceMatch(allResources)
		})

		Context("when given folders", func() {
			BeforeEach(func() {
				allResources = []Resource{
					{Filename: "folder-1", Mode: DefaultFolderPermissions},
					{Filename: "folder-2", Mode: DefaultFolderPermissions},
					{Filename: "folder-1/folder-3", Mode: DefaultFolderPermissions},
				}
			})

			It("does not send folders", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeCloudControllerClient.ResourceMatchCallCount()).To(Equal(0))
			})

			It("returns all folders [in order] in unmatchedResources", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(unmatchedResources).To(Equal(allResources))
			})
		})

		Context("when given files", func() {
			BeforeEach(func() {
				allResources = []Resource{
					{Filename: "file-1", Mode: 0744, Size: 11, SHA1: "some-sha-1"},
					{Filename: "file-2", Mode: 0744, Size: 0, SHA1: "some-sha-2"},
					{Filename: "file-3", Mode: 0744, Size: 13, SHA1: "some-sha-3"},
					{Filename: "file-4", Mode: 0744, Size: 14, SHA1: "some-sha-4"},
					{Filename: "file-5", Mode: 0744, Size: 15, SHA1: "some-sha-5"},
				}
			})

			It("sends non-zero sized files", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeCloudControllerClient.ResourceMatchCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.ResourceMatchArgsForCall(0)).To(ConsistOf(
					ccv2.Resource{Filename: "file-1", Mode: 0744, Size: 11, SHA1: "some-sha-1"},
					ccv2.Resource{Filename: "file-3", Mode: 0744, Size: 13, SHA1: "some-sha-3"},
					ccv2.Resource{Filename: "file-4", Mode: 0744, Size: 14, SHA1: "some-sha-4"},
					ccv2.Resource{Filename: "file-5", Mode: 0744, Size: 15, SHA1: "some-sha-5"},
				))
			})

			Context("when none of the files are matched", func() {
				It("returns all files [in order] in unmatchedResources", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(unmatchedResources).To(Equal(allResources))
				})
			})

			Context("when some files are matched", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.ResourceMatchReturns(
						[]ccv2.Resource{
							ccv2.Resource{Size: 14, SHA1: "some-sha-4"},
							ccv2.Resource{Size: 13, SHA1: "some-sha-3"},
						},
						ccv2.Warnings{"warnings-1", "warnings-2"},
						nil,
					)
				})

				It("returns all the unmatched files [in order] in unmatchedResources", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(unmatchedResources).To(ConsistOf(
						Resource{Filename: "file-1", Mode: 0744, Size: 11, SHA1: "some-sha-1"},
						Resource{Filename: "file-2", Mode: 0744, Size: 0, SHA1: "some-sha-2"},
						Resource{Filename: "file-5", Mode: 0744, Size: 15, SHA1: "some-sha-5"},
					))
				})

				It("returns all the matched files [in order] in matchedResources", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(matchedResources).To(ConsistOf(
						Resource{Filename: "file-3", Mode: 0744, Size: 13, SHA1: "some-sha-3"},
						Resource{Filename: "file-4", Mode: 0744, Size: 14, SHA1: "some-sha-4"},
					))
				})

				It("returns the warnings", func() {
					Expect(warnings).To(ConsistOf("warnings-1", "warnings-2"))
				})
			})
		})

		Context("when sending a large number of files/folders", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.ResourceMatchReturnsOnCall(
					0, nil, ccv2.Warnings{"warnings-1"}, nil,
				)
				fakeCloudControllerClient.ResourceMatchReturnsOnCall(
					1, nil, ccv2.Warnings{"warnings-2"}, nil,
				)

				allResources = []Resource{} // empties to prevent test pollution
				for i := 0; i < MaxResourceMatchChunkSize+2; i += 1 {
					allResources = append(allResources, Resource{Filename: "file", Mode: 0744, Size: 11, SHA1: "some-sha"})
				}
			})

			It("chunks the CC API calls by MaxResourceMatchChunkSize", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("warnings-1", "warnings-2"))

				Expect(fakeCloudControllerClient.ResourceMatchCallCount()).To(Equal(2))
				Expect(fakeCloudControllerClient.ResourceMatchArgsForCall(0)).To(HaveLen(MaxResourceMatchChunkSize))
				Expect(fakeCloudControllerClient.ResourceMatchArgsForCall(1)).To(HaveLen(2))
			})

		})

		Context("when the CC API returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("things are taking tooooooo long")
				fakeCloudControllerClient.ResourceMatchReturnsOnCall(
					0, nil, ccv2.Warnings{"warnings-1"}, nil,
				)
				fakeCloudControllerClient.ResourceMatchReturnsOnCall(
					1, nil, ccv2.Warnings{"warnings-2"}, expectedErr,
				)

				allResources = []Resource{} // empties to prevent test pollution
				for i := 0; i < MaxResourceMatchChunkSize+2; i += 1 {
					allResources = append(allResources, Resource{Filename: "file", Mode: 0744, Size: 11, SHA1: "some-sha"})
				}
			})

			It("returns all warnings and errors", func() {
				Expect(executeErr).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("warnings-1", "warnings-2"))
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
			tmpfile, err := ioutil.TempFile("", "zip-archive-resources")
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

		Context("when the files have not been changed since scanning them", func() {
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
				Expect(reader.File[4].Name).To(Equal("/tmpFile2"))

				expectFileContentsToEqual(reader.File[3], "why hello")
				expectFileContentsToEqual(reader.File[4], "Hello, Binky")

				for _, file := range reader.File {
					Expect(file.Method).To(Equal(zip.Deflate))
				}
			})
		})

		Context("when the files have changed since the scanning", func() {
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
				Expect(executeErr).To(Equal(FileChangedError{Filename: "/tmpFile3"}))
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

		Context("when the files have not been changed since scanning them", func() {
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
				Expect(reader.File[3].Name).To(Equal("tmpFile2"))
				Expect(reader.File[4].Name).To(Equal("tmpFile3"))

				expectFileContentsToEqual(reader.File[2], "why hello")
				expectFileContentsToEqual(reader.File[3], "Hello, Binky")
				expectFileContentsToEqual(reader.File[4], "Bananarama")

				for _, file := range reader.File {
					Expect(file.Method).To(Equal(zip.Deflate))
				}
			})
		})

		Context("when the files have changed since the scanning", func() {
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
				Expect(executeErr).To(Equal(FileChangedError{Filename: filepath.Join(srcDir, "tmpFile3")}))
			})
		})
	})
})

func expectFileContentsToEqual(file *zip.File, expectedContents string) {
	reader, err := file.Open()
	Expect(err).ToNot(HaveOccurred())
	defer reader.Close()

	body, err := ioutil.ReadAll(reader)
	Expect(err).ToNot(HaveOccurred())

	Expect(string(body)).To(Equal(expectedContents))
}
