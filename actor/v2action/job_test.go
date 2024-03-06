package v2action_test

import (
	"errors"
	"io"
	"strings"

	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Job Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v2actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil)
	})

	Describe("UploadApplicationPackage", func() {
		var (
			appGUID           string
			existingResources []Resource
			reader            io.Reader
			readerLength      int64

			job        Job
			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			appGUID = "some-app-guid"
			existingResources = []Resource{{Filename: "some-resource"}, {Filename: "another-resource"}}
			someString := "who reads these days"
			reader = strings.NewReader(someString)
			readerLength = int64(len([]byte(someString)))
		})

		JustBeforeEach(func() {
			job, warnings, executeErr = actor.UploadApplicationPackage(appGUID, existingResources, reader, readerLength)
		})

		When("the upload is successful", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.UploadApplicationPackageReturns(ccv2.Job{GUID: "some-job-guid"}, ccv2.Warnings{"upload-warning-1", "upload-warning-2"}, nil)
			})

			It("returns all warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("upload-warning-1", "upload-warning-2"))
				Expect(job).To(Equal(Job{GUID: "some-job-guid"}))

				Expect(fakeCloudControllerClient.UploadApplicationPackageCallCount()).To(Equal(1))
				passedAppGUID, passedExistingResources, passedReader, passedReaderLength := fakeCloudControllerClient.UploadApplicationPackageArgsForCall(0)
				Expect(passedAppGUID).To(Equal(appGUID))
				Expect(passedExistingResources).To(ConsistOf(ccv2.Resource{Filename: "some-resource"}, ccv2.Resource{Filename: "another-resource"}))
				Expect(passedReader).To(Equal(reader))
				Expect(passedReaderLength).To(Equal(readerLength))
			})
		})

		When("the upload returns an error", func() {
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
	})

	Describe("UploadDroplet", func() {
		var (
			appGUID       string
			droplet       io.Reader
			dropletLength int64

			job        Job
			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			appGUID = "some-app-guid"
			someString := "who reads these days"

			droplet = strings.NewReader(someString)
			dropletLength = int64(len([]byte(someString)))
		})

		JustBeforeEach(func() {
			job, warnings, executeErr = actor.UploadDroplet(appGUID, droplet, dropletLength)
		})

		When("the upload is successful", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.UploadDropletReturns(ccv2.Job{GUID: "some-job-guid"}, ccv2.Warnings{"upload-droplet-warning-1", "upload-droplet-warning-2"}, nil)
			})

			It("returns all warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("upload-droplet-warning-1", "upload-droplet-warning-2"))
				Expect(job).To(Equal(Job{GUID: "some-job-guid"}))

				Expect(fakeCloudControllerClient.UploadDropletCallCount()).To(Equal(1))
				passedAppGUID, passedDroplet, passedDropletLength := fakeCloudControllerClient.UploadDropletArgsForCall(0)
				Expect(passedAppGUID).To(Equal(appGUID))
				Expect(passedDroplet).To(Equal(droplet))
				Expect(passedDropletLength).To(Equal(dropletLength))
			})
		})

	})

	Describe("PollJob", func() {
		var (
			job        Job
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			warnings, executeErr = actor.PollJob(job)
		})

		When("the job polling is successful", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.PollJobReturns(ccv2.Warnings{"polling-warning"}, nil)
			})

			It("returns the warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("polling-warning"))

				Expect(fakeCloudControllerClient.PollJobCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.PollJobArgsForCall(0)).To(Equal(ccv2.Job(job)))
			})
		})

		When("polling errors", func() {
			var err error

			BeforeEach(func() {
				err = errors.New("some-error")
				fakeCloudControllerClient.PollJobReturns(ccv2.Warnings{"polling-warning"}, err)
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError(err))
				Expect(warnings).To(ConsistOf("polling-warning"))
			})
		})
	})
})
