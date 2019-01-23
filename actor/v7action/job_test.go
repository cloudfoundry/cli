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

		jobURL   ccv3.JobURL
		warnings Warnings
		err      error
	)

	BeforeEach(func() {
		actor, fakeCloudControllerClient, _, _, _ = NewTestActor()
		jobURL = ccv3.JobURL("http://example.com/the-job-url")
	})

	Describe("PollUploadBuildpackJob", func() {
		When("the cc client returns success", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.PollJobReturns(ccv3.Warnings{"some-warnings"}, nil)
			})

			It("returns success and warnings", func() {
				warnings, err = actor.PollUploadBuildpackJob(jobURL)

				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("some-warnings"))
				actualJobURL := fakeCloudControllerClient.PollJobArgsForCall(0)
				Expect(actualJobURL).To(Equal(jobURL))
			})
		})

		When("the cc client returns error", func() {
			var detailMsg string

			When("the cc client returns CF-BuildpackNameStackTaken", func() {
				BeforeEach(func() {
					detailMsg = "The buildpack name ruby_buildpack is already in use for the stack cflinuxfs2"
					fakeCloudControllerClient.PollJobReturns(ccv3.Warnings{"some-warnings"}, ccerror.V3JobFailedError{
						Code:   290000,
						Detail: detailMsg,
						Title:  "CF-BuildpackNameStackTaken",
					})
				})

				It("returns a converted error and warnings", func() {
					warnings, err = actor.PollUploadBuildpackJob(jobURL)

					Expect(err).To(Equal(ccerror.BuildpackAlreadyExistsForStackError{Message: detailMsg}))
					Expect(warnings).To(ConsistOf("some-warnings"))
					actualJobURL := fakeCloudControllerClient.PollJobArgsForCall(0)
					Expect(actualJobURL).To(Equal(jobURL))
				})
			})

			When("the cc client returns CF-BuildpackNameTaken", func() {
				BeforeEach(func() {
					detailMsg = "The buildpack name ruby_buildpack is already in use"
					fakeCloudControllerClient.PollJobReturns(ccv3.Warnings{"some-warnings"}, ccerror.V3JobFailedError{
						Code:   290001,
						Detail: detailMsg,
						Title:  "CF-BuildpackNameTaken",
					})
				})

				It("returns a converted error and warnings", func() {
					warnings, err = actor.PollUploadBuildpackJob(jobURL)

					Expect(err).To(Equal(ccerror.BuildpackNameTakenError{Message: detailMsg}))
					Expect(warnings).To(ConsistOf("some-warnings"))
					actualJobURL := fakeCloudControllerClient.PollJobArgsForCall(0)
					Expect(actualJobURL).To(Equal(jobURL))
				})
			})

			When("the cc client returns CF-BuildpackInvalid", func() {
				BeforeEach(func() {
					detailMsg = "The buildpack name ruby_buildpack is already in use"
					fakeCloudControllerClient.PollJobReturns(ccv3.Warnings{"some-warnings"}, ccerror.V3JobFailedError{
						Code:   290003,
						Detail: detailMsg,
						Title:  "CF-BuildpackInvalid",
					})
				})

				It("returns a converted error and warnings", func() {
					warnings, err = actor.PollUploadBuildpackJob(jobURL)

					Expect(err).To(Equal(ccerror.BuildpackAlreadyExistsWithoutStackError{Message: detailMsg}))
					Expect(warnings).To(ConsistOf("some-warnings"))
					actualJobURL := fakeCloudControllerClient.PollJobArgsForCall(0)
					Expect(actualJobURL).To(Equal(jobURL))
				})
			})
		})
	})
})
