package ccv3_test

import (
	"encoding/json"
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Relationship", func() {
	var (
		client *Client
	)

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("Relationship", func() {
		Describe("MarshalJSON", func() {
			Context("when the isolation segment is specified by name", func() {
				It("contains the name in the marshaled JSON", func() {
					body, err := json.Marshal(Relationship{GUID: "some-iso-guid"})
					expectedJSON := `{
					"data": {
						"guid": "some-iso-guid"
					}
				}`

					Expect(err).NotTo(HaveOccurred())
					Expect(body).To(MatchJSON(expectedJSON))
				})
			})

			Context("when the isolation segment is the empty string", func() {
				It("contains null in the marshaled JSON", func() {
					body, err := json.Marshal(Relationship{GUID: ""})
					expectedJSON := `{
					"data": null
				}`

					Expect(err).NotTo(HaveOccurred())
					Expect(body).To(MatchJSON(expectedJSON))
				})
			})
		})
	})

	Describe("AssignSpaceToIsolationSegment", func() {
		Context("when the assignment is successful", func() {
			BeforeEach(func() {
				response := `{
					"data": {
						"guid": "some-isolation-segment-guid"
					}
				}`

				requestBody := map[string]map[string]string{
					"data": {"guid": "some-iso-guid"},
				}
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPatch, "/v3/spaces/some-space-guid/relationships/isolation_segment"),
						VerifyJSONRepresenting(requestBody),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns all relationships and warnings", func() {
				relationship, warnings, err := client.AssignSpaceToIsolationSegment("some-space-guid", "some-iso-guid")
				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))
				Expect(relationship).To(Equal(Relationship{
					GUID: "some-isolation-segment-guid",
				}))
			})
		})
	})

	Describe("DeleteServiceInstanceRelationshipsSharedSpace", func() {
		var (
			serviceInstanceGUID string
			spaceGUID           string

			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			serviceInstanceGUID = "some-service-instance-guid"
			spaceGUID = "some-space-guid"
		})

		JustBeforeEach(func() {
			warnings, executeErr = client.DeleteServiceInstanceRelationshipsSharedSpace(serviceInstanceGUID, spaceGUID)
		})

		Context("when no errors occur deleting the shared space relationship", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodDelete, "/v3/service_instances/some-service-instance-guid/relationships/shared_spaces/some-space-guid"),
						RespondWith(http.StatusNoContent, "{}", http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("does not return any errors and returns all warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})

		Context("when an error occurs deleting the shared space relationship", func() {
			BeforeEach(func() {
				response := `{
						"errors": [
							{
								"code": 10008,
								"detail": "The request is semantically invalid: command presence",
								"title": "CF-UnprocessableEntity"
							}
						]
					}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodDelete, "/v3/service_instances/some-service-instance-guid/relationships/shared_spaces/some-space-guid"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the errors and all warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.V3UnexpectedResponseError{
					ResponseCode: http.StatusTeapot,
					V3ErrorResponse: ccerror.V3ErrorResponse{
						Errors: []ccerror.V3Error{
							{
								Code:   10008,
								Detail: "The request is semantically invalid: command presence",
								Title:  "CF-UnprocessableEntity",
							},
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("GetOrganizationDefaultIsolationSegment", func() {
		Context("when getting the isolation segment is successful", func() {
			BeforeEach(func() {
				response := `{
					"data": {
						"guid": "some-isolation-segment-guid"
					}
				}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/organizations/some-org-guid/relationships/default_isolation_segment"),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the relationship and warnings", func() {
				relationship, warnings, err := client.GetOrganizationDefaultIsolationSegment("some-org-guid")
				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))
				Expect(relationship).To(Equal(Relationship{
					GUID: "some-isolation-segment-guid",
				}))
			})
		})

		Context("when getting the isolation segment fails with an error", func() {
			BeforeEach(func() {
				response := `{
					"errors": [
						{
							"detail": "Organization not found",
							"title": "CF-ResourceNotFound",
							"code": 10010
						}
					]
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/organizations/some-org-guid/relationships/default_isolation_segment"),
						RespondWith(http.StatusNotFound, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns an error and warnings", func() {
				_, warnings, err := client.GetOrganizationDefaultIsolationSegment("some-org-guid")
				Expect(err).To(MatchError(ccerror.ResourceNotFoundError{
					Message: "Organization not found",
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("GetSpaceIsolationSegment", func() {
		Context("when getting the isolation segment is successful", func() {
			BeforeEach(func() {
				response := `{
					"data": {
						"guid": "some-isolation-segment-guid"
					}
				}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/spaces/some-space-guid/relationships/isolation_segment"),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the relationship and warnings", func() {
				relationship, warnings, err := client.GetSpaceIsolationSegment("some-space-guid")
				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))
				Expect(relationship).To(Equal(Relationship{
					GUID: "some-isolation-segment-guid",
				}))
			})
		})
	})

	Describe("PatchOrganizationDefaultIsolationSegment", func() {
		Context("when patching the default organization isolation segment with non-empty isolation segment guid", func() {
			BeforeEach(func() {
				expectedBody := `{
					"data": {
						"guid": "some-isolation-segment-guid"
					}
				}`
				responseBody := `{
					"data": {
						"guid": "some-isolation-segment-guid"
					}
				}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPatch, "/v3/organizations/some-org-guid/relationships/default_isolation_segment"),
						VerifyJSON(expectedBody),
						RespondWith(http.StatusOK, responseBody, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("patches the organization's default isolation segment", func() {
				relationship, warnings, err := client.PatchOrganizationDefaultIsolationSegment("some-org-guid", "some-isolation-segment-guid")
				Expect(relationship).To(Equal(Relationship{GUID: "some-isolation-segment-guid"}))
				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})

		Context("when patching the default organization isolation segment with empty isolation segment guid", func() {
			BeforeEach(func() {
				expectedBody := `{
					"data": null
				}`
				responseBody := `{
					"data": null
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPatch, "/v3/organizations/some-org-guid/relationships/default_isolation_segment"),
						VerifyJSON(expectedBody),
						RespondWith(http.StatusOK, responseBody, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("patches the organization's default isolation segment with nil guid", func() {
				relationship, warnings, err := client.PatchOrganizationDefaultIsolationSegment("some-org-guid", "")
				Expect(relationship).To(BeZero())
				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})

		Context("when patching the isolation segment fails with an error", func() {
			BeforeEach(func() {
				response := `{
					"errors": [
						{
							"detail": "Organization not found",
							"title": "CF-ResourceNotFound",
							"code": 10010
						}
					]
				}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPatch, "/v3/organizations/some-org-guid/relationships/default_isolation_segment"),
						RespondWith(http.StatusNotFound, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the empty relationship, an error and warnings", func() {
				relationship, warnings, err := client.PatchOrganizationDefaultIsolationSegment("some-org-guid", "some-isolation-segment-guid")
				Expect(relationship).To(BeZero())
				Expect(err).To(MatchError(ccerror.ResourceNotFoundError{
					Message: "Organization not found",
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("RevokeIsolationSegmentFromOrganization", func() {
		Context("when relationship exists", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodDelete, "/v3/isolation_segments/segment-guid/relationships/organizations/org-guid"),
						RespondWith(http.StatusOK, "", http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("revoke the relationship", func() {
				warnings, err := client.RevokeIsolationSegmentFromOrganization("segment-guid", "org-guid")
				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))

				Expect(server.ReceivedRequests()).To(HaveLen(3))
			})
		})

		Context("when an error occurs", func() {
			BeforeEach(func() {
				response := `{
					"errors": [
						{
							"code": 10008,
							"detail": "The request is semantically invalid: command presence",
							"title": "CF-UnprocessableEntity"
						}
					]
				}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodDelete, "/v3/isolation_segments/segment-guid/relationships/organizations/org-guid"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and warnings", func() {
				warnings, err := client.RevokeIsolationSegmentFromOrganization("segment-guid", "org-guid")
				Expect(err).To(MatchError(ccerror.V3UnexpectedResponseError{
					ResponseCode: http.StatusTeapot,
					V3ErrorResponse: ccerror.V3ErrorResponse{
						Errors: []ccerror.V3Error{
							{
								Code:   10008,
								Detail: "The request is semantically invalid: command presence",
								Title:  "CF-UnprocessableEntity",
							},
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("SetApplicationDroplet", func() {
		Context("it sets the droplet", func() {
			BeforeEach(func() {
				response := `
{
  "data": {
    "guid": "some-droplet-guid"
  },
  "links": {
    "self": {
      "href": "https://api.example.org/v3/apps/some-app-guid/relationships/current_droplet"
    },
    "related": {
      "href": "https://api.example.org/v3/apps/some-app-guid/droplets/current"
    }
  }
}`
				requestBody := map[string]interface{}{
					"data": map[string]string{
						"guid": "some-droplet-guid",
					},
				}

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPatch, "/v3/apps/some-app-guid/relationships/current_droplet"),
						VerifyJSONRepresenting(requestBody),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns warnings and no error", func() {
				relationship, warnings, err := client.SetApplicationDroplet("some-app-guid", "some-droplet-guid")
				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))
				Expect(relationship.GUID).To(Equal("some-droplet-guid"))
			})
		})

		Context("when the CC returns an error", func() {
			BeforeEach(func() {
				response := `{
  "errors": [
    {
      "code": 10008,
      "detail": "The request is semantically invalid: command presence",
      "title": "CF-UnprocessableEntity"
    },
    {
      "code": 10010,
      "detail": "App not found",
      "title": "CF-ResourceNotFound"
    }
  ]
}`
				requestBody := map[string]interface{}{
					"data": map[string]string{
						"guid": "some-droplet-guid",
					},
				}

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPatch, "/v3/apps/no-such-app-guid/relationships/current_droplet"),
						VerifyJSONRepresenting(requestBody),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)

			})

			It("returns the error and all warnings", func() {
				_, warnings, err := client.SetApplicationDroplet("no-such-app-guid", "some-droplet-guid")
				Expect(err).To(MatchError(ccerror.V3UnexpectedResponseError{
					ResponseCode: http.StatusTeapot,
					V3ErrorResponse: ccerror.V3ErrorResponse{
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
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})
})
