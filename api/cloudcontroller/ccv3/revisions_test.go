package ccv3_test

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/ccv3fakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
)

var _ = Describe("Revisions", func() {
	var (
		client    *Client
		requester *ccv3fakes.FakeRequester
	)

	BeforeEach(func() {
		requester = new(ccv3fakes.FakeRequester)
		client, _ = NewFakeRequesterTestClient(requester)
	})

	Describe("GetApplicationRevisions", func() {
		var (
			revisions  []Revision
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			revisions, warnings, executeErr = client.GetApplicationRevisions("some-app-guid")
		})

		When("the cloud controller returns errors and warnings", func() {
			BeforeEach(func() {
				errors := []ccerror.V3Error{
					{
						Code:   10008,
						Detail: "The request is semantically invalid: command presence",
						Title:  "CF-UnprocessableEntity",
					},
					{
						Code:   10010,
						Detail: "App not found",
						Title:  "CF-ResourceNotFound",
					},
				}

				requester.MakeListRequestReturns(
					IncludedResources{},
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
						{
							Code:   10010,
							Detail: "App not found",
							Title:  "CF-ResourceNotFound",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})

		When("applications exist", func() {
			BeforeEach(func() {
				requester.MakeListRequestCalls(func(requestParams RequestParams) (IncludedResources, Warnings, error) {
					err := requestParams.AppendToList(Revision{GUID: "app-guid-1"})
					Expect(err).NotTo(HaveOccurred())
					return IncludedResources{}, Warnings{"this is a warning", "this is another warning"}, nil
				})
			})

			It("returns the revisions and all warnings", func() {
				Expect(requester.MakeListRequestCallCount()).To(Equal(1))
				actualParams := requester.MakeListRequestArgsForCall(0)
				Expect(actualParams.RequestName).To(Equal(internal.GetApplicationRevisionsRequest))

				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning", "this is another warning"))

				Expect(revisions).To(ConsistOf([]Revision{{GUID: "app-guid-1"}}))
			})
		})

	})
})
