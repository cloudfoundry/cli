package ccv2_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
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

	Describe("GetOrganizations", func() {
		Context("when no errors are encountered", func() {
			Context("when results are paginated", func() {
				BeforeEach(func() {
					response1 := `{
					"next_url": "/v2/organizations?q=some-query:some-value&page=2",
					"resources": [
						{
							"metadata": {
								"guid": "org-guid-1"
							}
						},
						{
							"metadata": {
								"guid": "org-guid-2"
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
							}
						},
						{
							"metadata": {
								"guid": "org-guid-4"
							}
						}
					]
				}`
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v2/organizations", "q=some-query:some-value"),
							RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"warning-1"}}),
						))
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v2/organizations", "q=some-query:some-value&page=2"),
							RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"warning-2"}}),
						))
				})

				It("returns paginated results and all warnings", func() {
					orgs, warnings, err := client.GetOrganizations([]Query{{
						Filter:   "some-query",
						Operator: EqualOperator,
						Value:    "some-value",
					}})

					Expect(err).NotTo(HaveOccurred())
					Expect(orgs).To(Equal([]Organization{
						{GUID: "org-guid-1"},
						{GUID: "org-guid-2"},
						{GUID: "org-guid-3"},
						{GUID: "org-guid-4"},
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
						VerifyRequest(http.MethodGet, "/v2/organizations"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					))
			})

			It("returns an error and all warnings", func() {
				_, warnings, err := client.GetOrganizations(nil)

				Expect(err).To(MatchError(UnexpectedResponseError{
					ResponseCode: http.StatusTeapot,
					CCErrorResponse: CCErrorResponse{
						Code:        10001,
						Description: "Some Error",
						ErrorCode:   "CF-SomeError",
					},
				}))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})
	})

	Describe("DeleteOrganization", func() {
		Context("when no errors are encountered", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodDelete, "/v2/organizations/some-org-guid", "recursive=true&async=true"),
						RespondWith(http.StatusNoContent, "{}", http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					))
			})

			It("deletes the org and returns all warnings", func() {
				warnings, err := client.DeleteOrganization("some-org-guid")

				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf(Warnings{"warning-1", "warning-2"}))
			})
		})

		Context("when an error is encountered", func() {
			BeforeEach(func() {
				response := `{
  "code": 30003,
  "description": "The organization could not be found: some-org-guid",
  "error_code": "CF-OrganizationNotFound"
}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodDelete, "/v2/organizations/some-org-guid", "recursive=true&async=true"),
						RespondWith(http.StatusNotFound, response, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					))
			})

			It("returns an error and all warnings", func() {
				warnings, err := client.DeleteOrganization("some-org-guid")

				Expect(err).To(MatchError(cloudcontroller.ResourceNotFoundError{
					Message: "The organization could not be found: some-org-guid",
				}))
				Expect(warnings).To(ConsistOf(Warnings{"warning-1", "warning-2"}))
			})
		})
	})
})
