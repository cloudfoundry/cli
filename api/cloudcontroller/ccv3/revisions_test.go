package ccv3_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/ccv3fakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
			query      Query
			revisions  []resources.Revision
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			revisions, warnings, executeErr = client.GetApplicationRevisions("some-app-guid", query)
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
					err := requestParams.AppendToList(resources.Revision{GUID: "app-guid-1"})
					Expect(err).NotTo(HaveOccurred())
					return IncludedResources{}, Warnings{"this is a warning", "this is another warning"}, nil
				})
				query = Query{Key: OrderBy, Values: []string{"-created_at"}}
			})

			It("returns the revisions and all warnings", func() {
				Expect(requester.MakeListRequestCallCount()).To(Equal(1))
				actualParams := requester.MakeListRequestArgsForCall(0)
				Expect(actualParams.RequestName).To(Equal(internal.GetApplicationRevisionsRequest))
				Expect(actualParams.Query).To(Equal([]Query{query}))

				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning", "this is another warning"))

				Expect(revisions).To(ConsistOf([]resources.Revision{{GUID: "app-guid-1"}}))
			})
		})
	})

	Describe("GetApplicationRevisionsDeployed", func() {
		var (
			revisions  []resources.Revision
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			revisions, warnings, executeErr = client.GetApplicationRevisionsDeployed("some-app-guid")
		})

		When("applications exist", func() {
			BeforeEach(func() {
				requester.MakeListRequestCalls(func(requestParams RequestParams) (IncludedResources, Warnings, error) {
					err := requestParams.AppendToList(resources.Revision{GUID: "app-guid-1"})
					Expect(err).NotTo(HaveOccurred())
					return IncludedResources{}, Warnings{"this is a warning", "this is another warning"}, nil
				})
			})

			It("returns the revisions and all warnings", func() {
				Expect(requester.MakeListRequestCallCount()).To(Equal(1))
				actualParams := requester.MakeListRequestArgsForCall(0)
				Expect(actualParams.RequestName).To(Equal(internal.GetApplicationRevisionsDeployedRequest))

				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning", "this is another warning"))

				Expect(revisions).To(ConsistOf([]resources.Revision{{GUID: "app-guid-1"}}))
			})
		})
	})

	Describe("GetEnvironmentVariablesByURL", func() {
		var (
			warnings             Warnings
			executeErr           error
			environmentVariables resources.EnvironmentVariables
		)

		JustBeforeEach(func() {
			environmentVariables, warnings, executeErr = client.GetEnvironmentVariablesByURL("url")
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

				requester.MakeRequestReturns(
					"url",
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

		When("revision exist", func() {
			BeforeEach(func() {
				requester.MakeRequestCalls(func(requestParams RequestParams) (JobURL, Warnings, error) {
					(*requestParams.ResponseBody.(*resources.EnvironmentVariables))["foo"] = *types.NewFilteredString("bar")
					return "url", Warnings{"this is a warning"}, nil
				})
			})

			It("returns the environment variables and all warnings", func() {
				Expect(requester.MakeRequestCallCount()).To(Equal(1))
				actualParams := requester.MakeRequestArgsForCall(0)
				Expect(actualParams.URL).To(Equal("url"))

				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))

				Expect(len(environmentVariables)).To(Equal(1))
				Expect(environmentVariables["foo"].Value).To(Equal("bar"))
			})
		})
	})
})
