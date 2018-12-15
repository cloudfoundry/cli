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

var _ = Describe("Space", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("CreateSpace", func() {
		var (
			space      Space
			warnings   Warnings
			executeErr error
		)

		When("The response returns a valid space", func() {
			JustBeforeEach(func() {
				space, warnings, executeErr = client.CreateSpace("some-space", "some-org-guid")
			})

			BeforeEach(func() {
				response := `{
			  "metadata": {
		      "guid": "some-space-guid",
              "url": "/v2/spaces/some-space-guid"
			},
			  "entity": {
   				 "name": "some-space",
                 "organization_guid": "some-org-guid",
                 "space_quota_definition_guid": "some-quota-guid",
                 "allow_ssh": true
			  }
			}`

				requestBody := map[string]interface{}{
					"name":              "some-space",
					"organization_guid": "some-org-guid",
				}

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v2/spaces"),
						VerifyJSONRepresenting(requestBody),
						RespondWith(http.StatusCreated, response, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					))
			})

			It("return space and all warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(space).To(Equal(
					Space{
						GUID:                     "some-space-guid",
						OrganizationGUID:         "some-org-guid",
						Name:                     "some-space",
						SpaceQuotaDefinitionGUID: "some-quota-guid",
						AllowSSH:                 true,
					},
				))
				Expect(warnings).To(ConsistOf("warning-1"))
			})
		})

		When("The user is not authorized to create a space", func() {
			JustBeforeEach(func() {
				space, warnings, executeErr = client.CreateSpace("some-space", "some-org-guid")
			})

			BeforeEach(func() {
				response := `{
								"code": 10003,
								"description": "You are not authorized to perform the requested action"
							}`

				requestBody := map[string]interface{}{
					"name":              "some-space",
					"organization_guid": "some-org-guid",
				}

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v2/spaces"),
						VerifyJSONRepresenting(requestBody),
						RespondWith(http.StatusForbidden, response, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					))
			})

			It("return empty space with the errors and warnings", func() {
				Expect(space).To(Equal(
					Space{},
				))
				Expect(warnings).To(ConsistOf("warning-1"))
				Expect(executeErr).To(MatchError(ccerror.ForbiddenError{
					Message: "You are not authorized to perform the requested action",
				}))
			})
		})
	})

	Describe("DeleteSpace", func() {
		When("no errors are encountered", func() {
			BeforeEach(func() {
				jsonResponse := `{
				"metadata": {
					"guid": "job-guid",
					"created_at": "2016-06-08T16:41:27Z",
					"url": "/v2/jobs/job-guid"
				},
				"entity": {
					"guid": "job-guid",
					"status": "queued"
				}
			}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodDelete, "/v2/spaces/some-space-guid", "recursive=true&async=true"),
						RespondWith(http.StatusAccepted, jsonResponse, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					))
			})

			It("deletes the Space and returns all warnings", func() {
				job, warnings, err := client.DeleteSpace("some-space-guid")

				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf(Warnings{"warning-1", "warning-2"}))
				Expect(job.GUID).To(Equal("job-guid"))
				Expect(job.Status).To(Equal(constant.JobStatusQueued))
			})
		})

		When("an error is encountered", func() {
			BeforeEach(func() {
				response := `{
"code": 30003,
"description": "The Space could not be found: some-space-guid",
"error_code": "CF-SpaceNotFound"
}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodDelete, "/v2/spaces/some-space-guid", "recursive=true&async=true"),
						RespondWith(http.StatusNotFound, response, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					))
			})

			It("returns an error and all warnings", func() {
				_, warnings, err := client.DeleteSpace("some-space-guid")

				Expect(err).To(MatchError(ccerror.ResourceNotFoundError{
					Message: "The Space could not be found: some-space-guid",
				}))
				Expect(warnings).To(ConsistOf(Warnings{"warning-1", "warning-2"}))
			})
		})
	})

	Describe("GetSpaces", func() {
		When("no errors are encountered", func() {
			When("results are paginated", func() {
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
					spaces, warnings, err := client.GetSpaces(Filter{
						Type:     "some-query",
						Operator: constant.EqualOperator,
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

		When("an error is encountered", func() {
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

	Describe("GetSecurityGroupSpaces", func() {
		When("no errors are encountered", func() {
			When("results are paginated", func() {
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
					spaces, warnings, err := client.GetSecurityGroupSpaces("security-group-guid")

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

		When("an error is encountered", func() {
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
				_, warnings, err := client.GetSecurityGroupSpaces("security-group-guid")

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

	Describe("GetSecurityGroupStagingSpaces", func() {
		When("no errors are encountered", func() {
			When("results are paginated", func() {
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
					spaces, warnings, err := client.GetSecurityGroupStagingSpaces("security-group-guid")

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

		When("an error is encountered", func() {
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
				_, warnings, err := client.GetSecurityGroupStagingSpaces("security-group-guid")

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

	Describe("UpdateSpaceDeveloper", func() {
		var (
			executeErr error
			warnings   Warnings
		)

		JustBeforeEach(func() {
			warnings, executeErr = client.UpdateSpaceDeveloper("some-space-guid", "some-uaa-id")
		})

		When("no errors are encountered", func() {
			BeforeEach(func() {
				jsonResponse := `{
				"metadata": {
					"guid": "some-space-guid"
				},
				"entity": {
					"name": "some-space-name"
				}
				}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPut, "/v2/spaces/some-space-guid/developers/some-uaa-id"),
						RespondWith(http.StatusCreated, jsonResponse, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					))
			})

			It("makes the user a space developer and returns warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf(Warnings{"warning-1", "warning-2"}))
			})
		})

		When("the server errors", func() {
			BeforeEach(func() {
				jsonResponse := `{
							"code": 10001,
							"description": "Some Error",
							"error_code": "CF-SomeError"
						  }`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPut, "/v2/spaces/some-space-guid/developers/some-uaa-id"),
						RespondWith(http.StatusTeapot, jsonResponse, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					))
			})

			It("returns the error and all warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.V2UnexpectedResponseError{
					ResponseCode: http.StatusTeapot,
					V2ErrorResponse: ccerror.V2ErrorResponse{
						Code:        10001,
						Description: "Some Error",
						ErrorCode:   "CF-SomeError",
					},
				}))
				Expect(warnings).To(ConsistOf(Warnings{"warning-1", "warning-2"}))
			})
		})
	})
	Describe("UpdateSpaceDeveloperByUsername", func() {
		var (
			executeErr error
			warnings   Warnings
		)

		JustBeforeEach(func() {
			warnings, executeErr = client.UpdateSpaceDeveloperByUsername("some-space-guid", "user@example.com")
		})

		When("no errors are encountered", func() {
			BeforeEach(func() {
				jsonResponse := `{
				"metadata": {
					"guid": "some-space-guid"
				},
				"entity": {
					"name": "some-space-name",
					"organization_guid": "some-org-guid",
					"space_quota_definition_guid": null,
					"allow_ssh": true
				}
				}`

				requestBody := map[string]interface{}{
					"username": "user@example.com",
				}

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPut, "/v2/spaces/some-space-guid/developers"),
						VerifyJSONRepresenting(requestBody),
						RespondWith(http.StatusAccepted, jsonResponse, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					))
			})

			It("makes the user a space developer and returns warnings", func() {
				Expect(server.ReceivedRequests()).To(HaveLen(2))
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf(Warnings{"warning-1", "warning-2"}))
			})
		})

		When("the server errors", func() {
			BeforeEach(func() {
				jsonResponse := `{
					"code": 10001,
					"description": "Some Error",
					"error_code": "CF-SomeError"
				  }`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPut, "/v2/spaces/some-space-guid/developers"),
						RespondWith(http.StatusTeapot, jsonResponse, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					))
			})

			It("returns the error and all warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.V2UnexpectedResponseError{
					ResponseCode: http.StatusTeapot,
					V2ErrorResponse: ccerror.V2ErrorResponse{
						Code:        10001,
						Description: "Some Error",
						ErrorCode:   "CF-SomeError",
					},
				}))
				Expect(warnings).To(ConsistOf(Warnings{"warning-1", "warning-2"}))
			})
		})
	})

	Describe("UpdateSpaceManager", func() {
		var (
			warnings   Warnings
			executeErr error
		)
		JustBeforeEach(func() {
			warnings, executeErr = client.UpdateSpaceManager("some-space-guid", "some-uaa-id")
		})

		When("no errors are encountered", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPut, "/v2/spaces/some-space-guid/managers/some-uaa-id"),
						RespondWith(http.StatusCreated, `{}`, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					))
			})

			It("makes the user a space developer and returns warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf(Warnings{"warning-1", "warning-2"}))
			})
		})

		When("an error is encountered", func() {
			BeforeEach(func() {
				jsonResponse := `{
					"code": 10001,
					"description": "Some Error",
					"error_code": "CF-SomeError"
				  }`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPut, "/v2/spaces/some-space-guid/managers/some-uaa-id"),
						RespondWith(http.StatusTeapot, jsonResponse, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					))
			})

			It("returns the error and warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.V2UnexpectedResponseError{
					ResponseCode: http.StatusTeapot,
					V2ErrorResponse: ccerror.V2ErrorResponse{
						Code:        10001,
						Description: "Some Error",
						ErrorCode:   "CF-SomeError",
					},
				}))
				Expect(warnings).To(ConsistOf(Warnings{"warning-1", "warning-2"}))
			})
		})
	})

	Describe("UpdateSpaceManagerByUsername", func() {
		When("no errors are encountered", func() {
			BeforeEach(func() {
				jsonResponse := `{
					"metadata": {
					  "guid": "some-space-guid"
					},
					"entity": {
					  "name": "some-space-name",
					  "organization_guid": "some-org-guid",
					  "space_quota_definition_guid": null,
					  "allow_ssh": true
					}
				  }`

				requestBody := map[string]interface{}{
					"username": "user@example.com",
				}

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPut, "/v2/spaces/some-space-guid/managers"),
						VerifyJSONRepresenting(requestBody),
						RespondWith(http.StatusAccepted, jsonResponse, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					))
			})

			It("grants the permission and returns all warnings", func() {
				warnings, err := client.UpdateSpaceManagerByUsername("some-space-guid", "user@example.com")

				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf(Warnings{"warning-1", "warning-2"}))
			})
		})

		When("an error is encountered", func() {
			BeforeEach(func() {
				jsonResponse := `{
					"code": 10001,
					"description": "Some Error",
					"error_code": "CF-SomeError"
				  }`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPut, "/v2/spaces/some-space-guid/managers"),
						RespondWith(http.StatusTeapot, jsonResponse, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					))
			})

			It("grants the permission and returns all warnings", func() {
				warnings, err := client.UpdateSpaceManagerByUsername("some-space-guid", "user@example.com")

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
