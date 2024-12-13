package ccv3_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/ccv3fakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Revision", func() {
	var (
		client    *Client
		requester *ccv3fakes.FakeRequester
	)

	BeforeEach(func() {
		requester = new(ccv3fakes.FakeRequester)
		client, _ = NewFakeRequesterTestClient(requester)
	})

	Describe("GetRevision", func() {
		var (
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			_, warnings, executeErr = client.GetRevision("some-revision-guid")
		})

		When("the cloud controller returns errors and warnings", func() {
			BeforeEach(func() {
				errors := []ccerror.V3Error{
					{
						Code:   10008,
						Detail: "The request is semantically invalid: command presence",
						Title:  "CF-UnprocessableEntity",
					},
				}

				requester.MakeRequestReturns(
					"fake-job-url",
					Warnings{"this is a warning"},
					ccerror.MultiError{ResponseCode: http.StatusTeapot, Errors: errors},
				)
			})

			It("returns the error and all warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.MultiError{
					ResponseCode: http.StatusTeapot,
					Errors: []ccerror.V3Error{
						{
							Code:   10008,
							Detail: "The request is semantically invalid: command presence",
							Title:  "CF-UnprocessableEntity",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})

		When("revision exists", func() {
			BeforeEach(func() {
				requester.MakeRequestCalls(func(requestParams RequestParams) (ccv3.JobURL, Warnings, error) {
					requestParams.URIParams = internal.Params{"revision_guid": "some-revision-guid"}
					return JobURL(""), Warnings{"this is a warning", "this is another warning"}, nil
				})
			})

			It("returns the revision and all warnings", func() {
				Expect(requester.MakeRequestCallCount()).To(Equal(1))
				actualParams := requester.MakeRequestArgsForCall(0)
				Expect(actualParams.RequestName).To(Equal(internal.GetRevisionRequest))

				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning", "this is another warning"))
				Expect(actualParams.URIParams).To(Equal(internal.Params{"revision_guid": "some-revision-guid"}))
			})
		})
	})
})
