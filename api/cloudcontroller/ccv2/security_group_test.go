package ccv2_test

import (
	"net/http"

	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Security Groups", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("AssociateSpaceWithSecurityGroup", func() {
		Context("when no errors are encountered", func() {
			BeforeEach(func() {
				response := `{}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPut, "/v2/security_groups/security-group-guid/spaces/space-guid"),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					))
			})

			It("returns all warnings", func() {
				warnings, err := client.AssociateSpaceWithSecurityGroup("security-group-guid", "space-guid")

				Expect(err).NotTo(HaveOccurred())
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
						VerifyRequest(http.MethodPut, "/v2/security_groups/security-group-guid/spaces/space-guid"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					))
			})

			It("returns an error and all warnings", func() {
				warnings, err := client.AssociateSpaceWithSecurityGroup("security-group-guid", "space-guid")

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

	Describe("GetSecurityGroups", func() {
		Context("when no errors are encountered", func() {
			Context("when results are paginated", func() {
				BeforeEach(func() {
					response1 := `{
						"next_url": "/v2/security_groups?q=some-query:some-value&page=2",
						"resources": [
							{
								"metadata": {
									"guid": "security-group-guid-1",
									"url": "/v2/security_groups/security-group-guid-1"
								},
								"entity": {
									"name": "security-group-1",
									"rules": [
									],
									"running_default": false,
									"staging_default": false,
									"spaces_url": "/v2/security_groups/security-group-guid-1/spaces"
								}
							}
						]
					}`
					response2 := `{
						"next_url": null,
						"resources": [
							{
								"metadata": {
									"guid": "security-group-guid-2",
									"url": "/v2/security_groups/security-group-guid-2"
								},
								"entity": {
									"name": "security-group-2",
									"rules": [
									],
									"running_default": false,
									"staging_default": false,
									"spaces_url": "/v2/security_groups/security-group-guid-2/spaces"
								}
							}
						]
					}`
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v2/security_groups", "q=some-query:some-value"),
							RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"warning-1"}}),
						))
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v2/security_groups", "q=some-query:some-value&page=2"),
							RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"warning-2"}}),
						))
				})

				It("returns paginated results and all warnings", func() {
					securityGroups, warnings, err := client.GetSecurityGroups([]Query{{
						Filter:   "some-query",
						Operator: EqualOperator,
						Value:    "some-value",
					}})

					Expect(err).NotTo(HaveOccurred())
					Expect(securityGroups).To(Equal([]SecurityGroup{
						{
							GUID: "security-group-guid-1",
							Name: "security-group-1",
						},
						{
							GUID: "security-group-guid-2",
							Name: "security-group-2",
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
						VerifyRequest(http.MethodGet, "/v2/security_groups"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					))
			})

			It("returns an error and all warnings", func() {
				_, warnings, err := client.GetSecurityGroups(nil)

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
})
