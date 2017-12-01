package v2action_test

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
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
				for i := 0; i < MaxResourceMatchChunkSize+2; i++ {
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
				for i := 0; i < MaxResourceMatchChunkSize+2; i++ {
					allResources = append(allResources, Resource{Filename: "file", Mode: 0744, Size: 11, SHA1: "some-sha"})
				}
			})

			It("returns all warnings and errors", func() {
				Expect(executeErr).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("warnings-1", "warnings-2"))
			})
		})
	})
})
