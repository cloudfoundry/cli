package v2action_test

import (
	"archive/zip"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/ykk"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Resource Actions", func() {
	var (
		actor                     Actor
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

		err = ioutil.WriteFile(filepath.Join(subDir, "tmpFile1"), []byte("why hello"), 0600)
		Expect(err).ToNot(HaveOccurred())

		err = ioutil.WriteFile(filepath.Join(srcDir, "tmpFile2"), []byte("Hello, Binky"), 0600)
		Expect(err).ToNot(HaveOccurred())

		err = ioutil.WriteFile(filepath.Join(srcDir, "tmpFile3"), []byte("Bananarama"), 0600)
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
					{Filename: "level1/level2/tmpFile1"},
					{Filename: "tmpFile2"},
					{Filename: "tmpFile3"},
				}))
		})
	})

	Describe("UploadApplicationPackage", func() {
		var (
			appGUID           string
			existingResources []Resource
			reader            io.Reader
			readerLength      int64
			warnings          Warnings
			executeErr        error
		)

		BeforeEach(func() {
			appGUID = "some-app-guid"
			existingResources = []Resource{{Filename: "some-resource"}, {Filename: "another-resource"}}
			someString := "who reads these days"
			reader = strings.NewReader(someString)
			readerLength = int64(len([]byte(someString)))
		})

		JustBeforeEach(func() {
			warnings, executeErr = actor.UploadApplicationPackage(appGUID, existingResources, reader, readerLength)
		})

		Context("when the upload returns an error", func() {
			var err error

			BeforeEach(func() {
				err = errors.New("some-error")
				fakeCloudControllerClient.UploadApplicationPackageReturns(ccv2.Job{}, ccv2.Warnings{"upload-warning-1", "upload-warning-2"}, err)
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError(err))
				Expect(warnings).To(ConsistOf("upload-warning-1", "upload-warning-2"))
			})
		})

		Context("when polling errors", func() {
			var err error

			BeforeEach(func() {
				fakeCloudControllerClient.UploadApplicationPackageReturns(ccv2.Job{}, ccv2.Warnings{"upload-warning-1", "upload-warning-2"}, nil)

				err = errors.New("some-error")
				fakeCloudControllerClient.PollJobReturns(ccv2.Warnings{"polling-warning"}, err)
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError(err))
				Expect(warnings).To(ConsistOf("upload-warning-1", "upload-warning-2", "polling-warning"))
			})
		})

		Context("when the upload is successful", func() {
			var returnedJob ccv2.Job

			BeforeEach(func() {
				returnedJob = ccv2.Job{GUID: "some-job-guid"}
				fakeCloudControllerClient.UploadApplicationPackageReturns(returnedJob, ccv2.Warnings{"upload-warning-1", "upload-warning-2"}, nil)
				fakeCloudControllerClient.PollJobReturns(ccv2.Warnings{"polling-warning"}, nil)
			})

			It("returns all warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("upload-warning-1", "upload-warning-2", "polling-warning"))

				Expect(fakeCloudControllerClient.UploadApplicationPackageCallCount()).To(Equal(1))
				passedAppGUID, passedExistingResources, passedReader, passedReaderLength := fakeCloudControllerClient.UploadApplicationPackageArgsForCall(0)
				Expect(passedAppGUID).To(Equal(appGUID))
				Expect(passedExistingResources).To(ConsistOf(ccv2.Resource{Filename: "some-resource"}, ccv2.Resource{Filename: "another-resource"}))
				Expect(passedReader).To(Equal(reader))
				Expect(passedReaderLength).To(Equal(readerLength))

				Expect(fakeCloudControllerClient.PollJobCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.PollJobArgsForCall(0)).To(Equal(returnedJob))
			})
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
})

func expectFileContentsToEqual(file *zip.File, expectedContents string) {
	reader, err := file.Open()
	Expect(err).ToNot(HaveOccurred())
	defer reader.Close()

	body, err := ioutil.ReadAll(reader)
	Expect(err).ToNot(HaveOccurred())

	Expect(string(body)).To(Equal(expectedContents))
}
