package ccv2_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Space", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("GetSpaces", func() {
		Context("when no errors are encountered", func() {
			Context("when results are paginated", func() {
				BeforeEach(func() {
					response1 := `{
						"next_url": "/v2/spaces?q=some-query:some-value&page=2&order-by=name",
						"resources": [
							{
								"metadata": {
									"guid": "space-guid-1"
								},
								"entity": {
									"name": "space-1",
									"allow_ssh": false,
									"space_quota_definition_guid": "some-space-quota-guid-1",
									"organization_guid": "org-guid-1"
								}
							},
							{
								"metadata": {
									"guid": "space-guid-2"
								},
								"entity": {
									"name": "space-2",
									"allow_ssh": true,
									"space_quota_definition_guid": "some-space-quota-guid-2",
									"organization_guid": "org-guid-2"
								}
							}
						]
					}`
					response2 := `{
						"next_url": null,
						"resources": [
							{
								"metadata": {
									"guid": "space-guid-3"
								},
								"entity": {
									"name": "space-3",
									"allow_ssh": false,
									"space_quota_definition_guid": "some-space-quota-guid-3",
									"organization_guid": "org-guid-3"
								}
							},
							{
								"metadata": {
									"guid": "space-guid-4"
								},
								"entity": {
									"name": "space-4",
									"allow_ssh": true,
									"space_quota_definition_guid": "some-space-quota-guid-4",
									"organization_guid": "org-guid-4"
								}
							}
						]
					}`
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v2/spaces", "q=some-query:some-value&order-by=name"),
							RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"warning-1"}}),
						))
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v2/spaces", "q=some-query:some-value&page=2&order-by=name"),
							RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"warning-2"}}),
						))
				})

				It("returns paginated results and all warnings", func() {
					spaces, warnings, err := client.GetSpaces(QQuery{
						Filter:   "some-query",
						Operator: EqualOperator,
						Values:   []string{"some-value"},
					})

					Expect(err).NotTo(HaveOccurred())
					Expect(spaces).To(Equal([]Space{
						{
							GUID:                     "space-guid-1",
							OrganizationGUID:         "org-guid-1",
							Name:                     "space-1",
							AllowSSH:                 false,
							SpaceQuotaDefinitionGUID: "some-space-quota-guid-1",
						},
						{
							GUID:                     "space-guid-2",
							OrganizationGUID:         "org-guid-2",
							Name:                     "space-2",
							AllowSSH:                 true,
							SpaceQuotaDefinitionGUID: "some-space-quota-guid-2",
						},
						{
							GUID:                     "space-guid-3",
							OrganizationGUID:         "org-guid-3",
							Name:                     "space-3",
							AllowSSH:                 false,
							SpaceQuotaDefinitionGUID: "some-space-quota-guid-3",
						},
						{
							GUID:                     "space-guid-4",
							OrganizationGUID:         "org-guid-4",
							Name:                     "space-4",
							AllowSSH:                 true,
							SpaceQuotaDefinitionGUID: "some-space-quota-guid-4",
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
						VerifyRequest(http.MethodGet, "/v2/spaces", "order-by=name"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					))
			})

			It("returns an error and all warnings", func() {
				_, warnings, err := client.GetSpaces()

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

	Describe("GetStagingSpacesBySecurityGroup", func() {
		Context("when no errors are encountered", func() {
			Context("when results are paginated", func() {
				BeforeEach(func() {
					response1 := `{
						"next_url": "/v2/security_groups/security-group-guid/staging_spaces?page=2",
						"resources": [
							{
								"metadata": {
									"guid": "space-guid-1"
								},
								"entity": {
									"name": "space-1",
									"allow_ssh": false,
									"space_quota_definition_guid": "some-space-quota-guid-1",
									"organization_guid": "org-guid-1"
								}
							},
							{
								"metadata": {
									"guid": "space-guid-2"
								},
								"entity": {
									"name": "space-2",
									"allow_ssh": true,
									"space_quota_definition_guid": "some-space-quota-guid-2",
									"organization_guid": "org-guid-2"
								}
							}
						]
					}`
					response2 := `{
						"next_url": null,
						"resources": [
							{
								"metadata": {
									"guid": "space-guid-3"
								},
								"entity": {
									"name": "space-3",
									"allow_ssh": false,
									"space_quota_definition_guid": "some-space-quota-guid-3",
									"organization_guid": "org-guid-3"
								}
							},
							{
								"metadata": {
									"guid": "space-guid-4"
								},
								"entity": {
									"name": "space-4",
									"allow_ssh": true,
									"space_quota_definition_guid": "some-space-quota-guid-4",
									"organization_guid": "org-guid-4"
								}
							}
						]
					}`
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v2/security_groups/security-group-guid/staging_spaces", ""),
							RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"warning-1"}}),
						))
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v2/security_groups/security-group-guid/staging_spaces", "page=2"),
							RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"warning-2"}}),
						))
				})

				It("returns paginated results and all warnings", func() {
					spaces, warnings, err := client.GetStagingSpacesBySecurityGroup("security-group-guid")

					Expect(err).NotTo(HaveOccurred())
					Expect(spaces).To(Equal([]Space{
						{
							GUID:                     "space-guid-1",
							OrganizationGUID:         "org-guid-1",
							Name:                     "space-1",
							AllowSSH:                 false,
							SpaceQuotaDefinitionGUID: "some-space-quota-guid-1",
						},
						{
							GUID:                     "space-guid-2",
							OrganizationGUID:         "org-guid-2",
							Name:                     "space-2",
							AllowSSH:                 true,
							SpaceQuotaDefinitionGUID: "some-space-quota-guid-2",
						},
						{
							GUID:                     "space-guid-3",
							OrganizationGUID:         "org-guid-3",
							Name:                     "space-3",
							AllowSSH:                 false,
							SpaceQuotaDefinitionGUID: "some-space-quota-guid-3",
						},
						{
							GUID:                     "space-guid-4",
							OrganizationGUID:         "org-guid-4",
							Name:                     "space-4",
							AllowSSH:                 true,
							SpaceQuotaDefinitionGUID: "some-space-quota-guid-4",
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
						VerifyRequest(http.MethodGet, "/v2/security_groups/security-group-guid/staging_spaces"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					))
			})

			It("returns an error and all warnings", func() {
				_, warnings, err := client.GetStagingSpacesBySecurityGroup("security-group-guid")

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

	Describe("GetRunningSpacesBySecurityGroup", func() {
		Context("when no errors are encountered", func() {
			Context("when results are paginated", func() {
				BeforeEach(func() {
					response1 := `{
						"next_url": "/v2/security_groups/security-group-guid/spaces?page=2",
						"resources": [
							{
								"metadata": {
									"guid": "space-guid-1"
								},
								"entity": {
									"name": "space-1",
									"allow_ssh": false,
									"space_quota_definition_guid": "some-space-quota-guid-1",
									"organization_guid": "org-guid-1"
								}
							},
							{
								"metadata": {
									"guid": "space-guid-2"
								},
								"entity": {
									"name": "space-2",
									"allow_ssh": true,
									"space_quota_definition_guid": "some-space-quota-guid-2",
									"organization_guid": "org-guid-2"
								}
							}
						]
					}`
					response2 := `{
						"next_url": null,
						"resources": [
							{
								"metadata": {
									"guid": "space-guid-3"
								},
								"entity": {
									"name": "space-3",
									"allow_ssh": false,
									"space_quota_definition_guid": "some-space-quota-guid-3",
									"organization_guid": "org-guid-3"
								}
							},
							{
								"metadata": {
									"guid": "space-guid-4"
								},
								"entity": {
									"name": "space-4",
									"allow_ssh": true,
									"space_quota_definition_guid": "some-space-quota-guid-4",
									"organization_guid": "org-guid-4"
								}
							}
						]
					}`
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v2/security_groups/security-group-guid/spaces", ""),
							RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"warning-1"}}),
						))
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v2/security_groups/security-group-guid/spaces", "page=2"),
							RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"warning-2"}}),
						))
				})

				It("returns paginated results and all warnings", func() {
					spaces, warnings, err := client.GetRunningSpacesBySecurityGroup("security-group-guid")

					Expect(err).NotTo(HaveOccurred())
					Expect(spaces).To(Equal([]Space{
						{
							GUID:                     "space-guid-1",
							OrganizationGUID:         "org-guid-1",
							Name:                     "space-1",
							AllowSSH:                 false,
							SpaceQuotaDefinitionGUID: "some-space-quota-guid-1",
						},
						{
							GUID:                     "space-guid-2",
							OrganizationGUID:         "org-guid-2",
							Name:                     "space-2",
							AllowSSH:                 true,
							SpaceQuotaDefinitionGUID: "some-space-quota-guid-2",
						},
						{
							GUID:                     "space-guid-3",
							OrganizationGUID:         "org-guid-3",
							Name:                     "space-3",
							AllowSSH:                 false,
							SpaceQuotaDefinitionGUID: "some-space-quota-guid-3",
						},
						{
							GUID:                     "space-guid-4",
							OrganizationGUID:         "org-guid-4",
							Name:                     "space-4",
							AllowSSH:                 true,
							SpaceQuotaDefinitionGUID: "some-space-quota-guid-4",
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
						VerifyRequest(http.MethodGet, "/v2/security_groups/security-group-guid/spaces"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					))
			})

			It("returns an error and all warnings", func() {
				_, warnings, err := client.GetRunningSpacesBySecurityGroup("security-group-guid")

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
