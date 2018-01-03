package ccv2_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Organization", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("GetOrganization", func() {
		Context("when the organization exists", func() {
			BeforeEach(func() {
				response := `{
					"metadata": {
						"guid": "some-org-guid"
					},
					"entity": {
						"name": "some-org",
						"quota_definition_guid": "some-quota-guid"
					}
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/organizations/some-org-guid"),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					))
			})

			It("returns the org and all warnings", func() {
				orgs, warnings, err := client.GetOrganization("some-org-guid")

				Expect(err).NotTo(HaveOccurred())
				Expect(orgs).To(Equal(
					Organization{
						GUID:                "some-org-guid",
						Name:                "some-org",
						QuotaDefinitionGUID: "some-quota-guid",
					},
				))
				Expect(warnings).To(ConsistOf("warning-1"))
			})
		})

		Context("when an error is encountered", func() {
			BeforeEach(func() {
				response := `{
					"code": 10001,
					"description": "Some Error",
					"error_code": "CF-SomeError"
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/organizations/some-org-guid"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					))
			})

			It("returns an error and all warnings", func() {
				_, warnings, err := client.GetOrganization("some-org-guid")

				Expect(err).To(MatchError(ccerror.V2UnexpectedResponseError{
					ResponseCode: http.StatusTeapot,
					V2ErrorResponse: ccerror.V2ErrorResponse{
						Code:        10001,
						Description: "Some Error",
						ErrorCode:   "CF-SomeError",
					},
				}))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})
	})

	Describe("GetOrganizations", func() {
		Context("when no errors are encountered", func() {
			Context("when results are paginated", func() {
				BeforeEach(func() {
					response1 := `{
					"next_url": "/v2/organizations?q=some-query:some-value&page=2&order-by=name",
					"resources": [
						{
							"metadata": {
								"guid": "org-guid-1"
							},
							"entity": {
								"name": "org-1",
								"quota_definition_guid": "some-quota-guid",
								"default_isolation_segment_guid": "some-default-isolation-segment-guid"
							}
						},
						{
							"metadata": {
								"guid": "org-guid-2"
							},
							"entity": {
								"name": "org-2",
								"quota_definition_guid": "some-quota-guid"
							}
						}
					]
				}`
					response2 := `{
					"next_url": null,
					"resources": [
						{
							"metadata": {
								"guid": "org-guid-3"
							},
							"entity": {
								"name": "org-3",
								"quota_definition_guid": "some-quota-guid"
							}
						},
						{
							"metadata": {
								"guid": "org-guid-4"
							},
							"entity": {
								"name": "org-4",
								"quota_definition_guid": "some-quota-guid"
							}
						}
					]
				}`
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v2/organizations", "q=some-query:some-value&order-by=name"),
							RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"warning-1"}}),
						))
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v2/organizations", "q=some-query:some-value&page=2&order-by=name"),
							RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"warning-2"}}),
						))
				})

				It("returns paginated results and all warnings", func() {
					orgs, warnings, err := client.GetOrganizations(QQuery{
						Filter:   "some-query",
						Operator: EqualOperator,
						Values:   []string{"some-value"},
					})

					Expect(err).NotTo(HaveOccurred())
					Expect(orgs).To(Equal([]Organization{
						{
							GUID:                        "org-guid-1",
							Name:                        "org-1",
							QuotaDefinitionGUID:         "some-quota-guid",
							DefaultIsolationSegmentGUID: "some-default-isolation-segment-guid",
						},
						{
							GUID:                        "org-guid-2",
							Name:                        "org-2",
							QuotaDefinitionGUID:         "some-quota-guid",
							DefaultIsolationSegmentGUID: "",
						},
						{
							GUID:                        "org-guid-3",
							Name:                        "org-3",
							QuotaDefinitionGUID:         "some-quota-guid",
							DefaultIsolationSegmentGUID: "",
						},
						{
							GUID:                        "org-guid-4",
							Name:                        "org-4",
							QuotaDefinitionGUID:         "some-quota-guid",
							DefaultIsolationSegmentGUID: "",
						},
					}))
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				})
			})
		})

		Context("when an error is encountered", func() {
			BeforeEach(func() {
				response := `{
  "code": 10001,
  "description": "Some Error",
  "error_code": "CF-SomeError"
}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/organizations", "order-by=name"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					))
			})

			It("returns an error and all warnings", func() {
				_, warnings, err := client.GetOrganizations()

				Expect(err).To(MatchError(ccerror.V2UnexpectedResponseError{
					ResponseCode: http.StatusTeapot,
					V2ErrorResponse: ccerror.V2ErrorResponse{
						Code:        10001,
						Description: "Some Error",
						ErrorCode:   "CF-SomeError",
					},
				}))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})
	})
})
