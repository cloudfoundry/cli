package v7action_test

import (
	"fmt"

	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
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
			DescribeTable("extracting actual error from JobFailedError",
				func(code int, expectedErrType error) {
					fakeCloudControllerClient.PollJobReturns(ccv3.Warnings{"some-warnings"}, ccerror.V3JobFailedError{
						Code:   int64(code),
						Detail: fmt.Sprintf("code %d", code),
					})
					warnings, err = actor.PollUploadBuildpackJob(jobURL)

					Expect(err).To(MatchError(expectedErrType))
					Expect(warnings).To(ConsistOf("some-warnings"))
					actualJobURL := fakeCloudControllerClient.PollJobArgsForCall(0)
					Expect(actualJobURL).To(Equal(jobURL))
				},
				Entry("BuildpackNameStackTaken", 290000, ccerror.BuildpackAlreadyExistsForStackError{Message: "code 290000"}),
				Entry("BuildpackInvalid", 290003, ccerror.BuildpackAlreadyExistsWithoutStackError{Message: "code 290003"}),
				Entry("BuildpackStacksDontMatch", 390011, ccerror.BuildpackStacksDontMatchError{Message: "code 390011"}),
				Entry("BuildpackStackDoesNotExist", 390012, ccerror.BuildpackStackDoesNotExistError{Message: "code 390012"}),
				Entry("BuildpackZipError", 390013, ccerror.BuildpackZipInvalidError{Message: "code 390013"}),
			)
		})
	})
})
