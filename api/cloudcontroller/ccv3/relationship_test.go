package ccv3_test

import (
	"encoding/json"
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
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

	Describe("Relationship.MarshalJSON", func() {
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
					"data": {
						"guid": null
					}
				}`

				Expect(err).NotTo(HaveOccurred())
				Expect(body).To(MatchJSON(expectedJSON))
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

	Describe("EntitleIsolationSegmentToOrganizations", func() {
		Context("when the delete is successful", func() {
			BeforeEach(func() {
				response := `{
					"data": [
						{
							"guid": "some-relationship-guid-1"
						},
						{
							"guid": "some-relationship-guid-2"
						}
					]
				}`

				requestBody := map[string][]map[string]string{
					"data": {{"guid": "org-guid-1"}, {"guid": "org-guid-2"}},
				}
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/isolation_segments/some-iso-guid/relationships/organizations"),
						VerifyJSONRepresenting(requestBody),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns all relationships and warnings", func() {
				relationships, warnings, err := client.EntitleIsolationSegmentToOrganizations("some-iso-guid", []string{"org-guid-1", "org-guid-2"})
				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))
				Expect(relationships).To(Equal(RelationshipList{
					GUIDs: []string{"some-relationship-guid-1", "some-relationship-guid-2"},
				}))
			})
		})

		Context("when the cloud controller returns errors and warnings", func() {
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
						VerifyRequest(http.MethodPost, "/v3/isolation_segments/some-iso-guid/relationships/organizations"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and all warnings", func() {
				_, warnings, err := client.EntitleIsolationSegmentToOrganizations("some-iso-guid", []string{"org-guid-1", "org-guid-2"})
				Expect(err).To(MatchError(ccerror.V3UnexpectedResponseError{
					ResponseCode: http.StatusTeapot,
					V3ErrorResponse: ccerror.V3ErrorResponse{
						[]ccerror.V3Error{
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

	Describe("RevokeIsolationSegmentFromOrganization ", func() {
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
	})

	Context("when relationship exists", func() {
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

		It("revoke the relationship", func() {
			warnings, err := client.RevokeIsolationSegmentFromOrganization("segment-guid", "org-guid")
			Expect(err).To(MatchError(ccerror.V3UnexpectedResponseError{
				ResponseCode: http.StatusTeapot,
				V3ErrorResponse: ccerror.V3ErrorResponse{
					[]ccerror.V3Error{
						{
							Code:   10008,
							Detail: "The request is semantically invalid: command presence",
							Title:  "CF-UnprocessableEntity",
						},
					},
				},
			}))
			Expect(warnings).To(ConsistOf("this is a warning"))

			Expect(server.ReceivedRequests()).To(HaveLen(3))
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

			It("returns the relationship and warnings", func() {
				_, warnings, err := client.GetOrganizationDefaultIsolationSegment("some-org-guid")
				Expect(err).To(MatchError(cloudcontroller.ResourceNotFoundError{
					Message: "Organization not found",
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})
})
