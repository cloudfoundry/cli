package v7action_test

import (
	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Job Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		actor, fakeCloudControllerClient, _, _, _, _, _ = NewTestActor()
	})

	Describe("PollUploadBuildpackJob", func() {
		var (
			jobURL   ccv3.JobURL
			warnings Warnings
			err      error
		)

		BeforeEach(func() {
			jobURL = ccv3.JobURL("http://example.com/the-job-url")
		})

		JustBeforeEach(func() {
			warnings, err = actor.PollUploadBuildpackJob(jobURL)
		})

		When("the cc client returns success", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.PollJobReturns(ccv3.Warnings{"some-warnings"}, nil)
			})

			It("returns success and warnings", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("some-warnings"))
			})

			It("calls poll job", func() {
				Expect(fakeCloudControllerClient.PollJobCallCount()).To(Equal(1))
				actualJobURL := fakeCloudControllerClient.PollJobArgsForCall(0)
				Expect(actualJobURL).To(Equal(jobURL))
			})
		})

		When("the cc client returns error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.PollJobReturns(ccv3.Warnings{"some-warnings"}, ccerror.V3JobFailedError{Detail: "some-err"})
			})

			It("returns the errors and warnings", func() {
				warnings, err = actor.PollUploadBuildpackJob(jobURL)

				Expect(err).To(MatchError(ccerror.V3JobFailedError{Detail: "some-err"}))
				Expect(warnings).To(ConsistOf("some-warnings"))
			})
		})
	})
})
