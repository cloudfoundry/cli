package ccv2_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("OrganizationQuota", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("GetOrganizationQuota", func() {

		When("getting the organization quota does not return an error", func() {
			BeforeEach(func() {
				response := `{
				"metadata": {
					"guid": "some-org-quota-guid"
				},
				"entity": {
					"name": "some-org-quota"
				}
			}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/quota_definitions/some-org-quota-guid"),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					),
				)
			})

			It("returns the organization quota", func() {
				orgQuota, warnings, err := client.GetOrganizationQuota("some-org-quota-guid")
				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(Equal(Warnings{"warning-1"}))
				Expect(orgQuota).To(Equal(OrganizationQuota{
					GUID: "some-org-quota-guid",
					Name: "some-org-quota",
				}))
			})
		})

		When("the organization quota returns an error", func() {
			BeforeEach(func() {
				response := `{
				  "description": "Quota Definition could not be found: some-org-quota-guid",
				  "error_code": "CF-QuotaDefinitionNotFound",
				  "code": 240001
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/quota_definitions/some-org-quota-guid"),
						RespondWith(http.StatusNotFound, response, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					),
				)
			})

			It("returns the error", func() {
				_, warnings, err := client.GetOrganizationQuota("some-org-quota-guid")
				Expect(err).To(MatchError(ccerror.ResourceNotFoundError{
					Message: "Quota Definition could not be found: some-org-quota-guid",
				}))
				Expect(warnings).To(Equal(Warnings{"warning-1"}))
			})
		})

	})

	Describe("GetOrganizationQuotas", func() {
		var (
			quotas     []OrganizationQuota
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			quotas, warnings, executeErr = client.GetOrganizationQuotas(Filter{
				Type:     "some-query",
				Operator: constant.EqualOperator,
				Values:   []string{"some-value"},
			})
		})

		When("listing the quotas succeeds", func() {
			BeforeEach(func() {
				response1 := `{
					"next_url": "/v2/quota_definitions?q=some-query:some-value&page=2",
					"resources": [
						{
							"metadata": {
								"guid": "some-org-quota-guid-1"
							},
							"entity": {
								"name": "some-quota-name-1"
							}
						},
						{
							"metadata": {
								"guid": "some-org-quota-guid-2"
							},
							"entity": {
								"name": "some-quota-name-2"
							}
						}
					]
				}`

				response2 := `{
					"next_url": null,
					"resources": [
						{
							"metadata": {
								"guid": "some-org-quota-guid-3"
							},
							"entity": {
								"name": "some-quota-name-3"
							}
						},
						{
							"metadata": {
								"guid": "some-org-quota-guid-4"
							},
							"entity": {
								"name": "some-quota-name-4"
							}
						}
					]
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/quota_definitions", "q=some-query:some-value"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					),
				)
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/quota_definitions", "q=some-query:some-value&page=2"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"warning-2"}}),
					),
				)
			})

			It("returns paginated results and all warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf(Warnings{"warning-1", "warning-2"}))
				Expect(quotas).To(Equal([]OrganizationQuota{
					{
						GUID: "some-org-quota-guid-1",
						Name: "some-quota-name-1",
					},
					{
						GUID: "some-org-quota-guid-2",
						Name: "some-quota-name-2",
					},
					{
						GUID: "some-org-quota-guid-3",
						Name: "some-quota-name-3",
					},
					{
						GUID: "some-org-quota-guid-4",
						Name: "some-quota-name-4",
					},
				}))

			})
		})

		Context("when the server errors", func() {
			BeforeEach(func() {
				response := `{
					"code": 10001,
					"description": "Some Error",
					"error_code": "CF-SomeError"
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/quota_definitions", "q=some-query:some-value"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					),
				)
			})

			It("returns warnings and errors", func() {
				Expect(executeErr).To(MatchError(ccerror.V2UnexpectedResponseError{
					ResponseCode: http.StatusTeapot,
					V2ErrorResponse: ccerror.V2ErrorResponse{
						Code:        10001,
						Description: "Some Error",
						ErrorCode:   "CF-SomeError",
					},
				}))
				Expect(warnings).To(Equal(Warnings{"warning-1"}))
			})
		})
	})
})
