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

var _ = Describe("Organization", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("CreateOrganization", func() {
		var (
			org        Organization
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			org, warnings, executeErr = client.CreateOrganization("some-org", "some-quota-guid")
		})

		When("the organization exists", func() {
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
				requestBody := map[string]interface{}{
					"name":                  "some-org",
					"quota_definition_guid": "some-quota-guid",
				}

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v2/organizations"),
						VerifyJSONRepresenting(requestBody),
						RespondWith(http.StatusCreated, response, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					))
			})

			It("returns the org and all warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(org).To(Equal(
					Organization{
						GUID:                "some-org-guid",
						Name:                "some-org",
						QuotaDefinitionGUID: "some-quota-guid",
					},
				))
				Expect(warnings).To(ConsistOf("warning-1"))
			})
		})
	})

	Describe("DeleteOrganization", func() {
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
						VerifyRequest(http.MethodDelete, "/v2/organizations/some-organization-guid", "recursive=true&async=true"),
						RespondWith(http.StatusAccepted, jsonResponse, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					))
			})

			It("deletes the Organization and returns all warnings", func() {
				job, warnings, err := client.DeleteOrganization("some-organization-guid")

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
"description": "The Organization could not be found: some-organization-guid",
"error_code": "CF-OrganizationNotFound"
}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodDelete, "/v2/organizations/some-organization-guid", "recursive=true&async=true"),
						RespondWith(http.StatusNotFound, response, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					))
			})

			It("returns an error and all warnings", func() {
				_, warnings, err := client.DeleteOrganization("some-organization-guid")

				Expect(err).To(MatchError(ccerror.ResourceNotFoundError{
					Message: "The Organization could not be found: some-organization-guid",
				}))
				Expect(warnings).To(ConsistOf(Warnings{"warning-1", "warning-2"}))
			})
		})
	})

	Describe("GetOrganization", func() {
		When("the organization exists", func() {
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

		When("an error is encountered", func() {
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
		When("no errors are encountered", func() {
			When("results are paginated", func() {
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
					orgs, warnings, err := client.GetOrganizations(Filter{
						Type:     "some-query",
						Operator: constant.EqualOperator,
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

		When("an error is encountered", func() {
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

	Describe("UpdateOrganizationManagerByUsername", func() {
		Context("when the organization exists", func() {
			var (
				warnings Warnings
				err      error
			)

			BeforeEach(func() {
				expectedRequest := `{
					"username": "some-user"
				}`

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
						VerifyRequest(http.MethodPut, "/v2/organizations/some-org-guid/managers"),
						VerifyJSON(expectedRequest),
						RespondWith(http.StatusCreated, response, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					))
			})

			JustBeforeEach(func() {
				warnings, err = client.UpdateOrganizationManagerByUsername("some-org-guid", "some-user")
			})

			It("returns warnings", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning-1"))
			})
		})
	})

	Describe("UpdateOrganizationManager", func() {
		var (
			warnings Warnings
			err      error
		)

		JustBeforeEach(func() {
			warnings, err = client.UpdateOrganizationManager("some-org-guid", "some-manager-guid")
		})

		When("the organization exists", func() {
			BeforeEach(func() {
				response := `{
					"metadata": {
						"guid": "some-org-guid"
					},
					"entity": {
						"name": "some-org",
					}
				}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPut, "/v2/organizations/some-org-guid/managers/some-manager-guid"),
						RespondWith(http.StatusCreated, response, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					))
			})

			It("returns warnings", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning-1"))
			})
		})

		When("the server returns an error", func() {
			BeforeEach(func() {
				response := `{
					"code": 10001,
					"description": "Some Error",
					"error_code": "CF-SomeError"
				}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPut, "/v2/organizations/some-org-guid/managers/some-manager-guid"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					))
			})

			It("returns the error and any warnings", func() {
				Expect(err).To(MatchError(ccerror.V2UnexpectedResponseError{
					ResponseCode: http.StatusTeapot,
					V2ErrorResponse: ccerror.V2ErrorResponse{
						Code:        10001,
						Description: "Some Error",
						ErrorCode:   "CF-SomeError",
					},
				}))
				Expect(warnings).To(ConsistOf("warning-1"))
			})
		})
	})

	Describe("UpdateOrganizationUserByUsername", func() {
		var (
			warnings Warnings
			err      error
		)

		JustBeforeEach(func() {
			warnings, err = client.UpdateOrganizationUserByUsername("some-org-guid", "some-user")
		})

		When("the request succeeds", func() {
			BeforeEach(func() {
				expectedRequest := `{
					"username": "some-user"
				}`

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
						VerifyRequest(http.MethodPut, "/v2/organizations/some-org-guid/users"),
						VerifyJSON(expectedRequest),
						RespondWith(http.StatusCreated, response, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					))
			})

			It("returns warnings", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning-1"))
			})
		})

		When("the server returns an error", func() {
			BeforeEach(func() {
				response := `{
					"code": 10001,
					"description": "Some Error",
					"error_code": "CF-SomeError"
				}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPut, "/v2/organizations/some-org-guid/users"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					))
			})

			It("returns warnings", func() {
				Expect(warnings).To(ConsistOf("warning-1"))
			})

			It("returns the error", func() {
				Expect(err).To(MatchError(ccerror.V2UnexpectedResponseError{
					ResponseCode: http.StatusTeapot,
					V2ErrorResponse: ccerror.V2ErrorResponse{
						Code:        10001,
						Description: "Some Error",
						ErrorCode:   "CF-SomeError",
					},
				}))
			})
		})
	})
})
