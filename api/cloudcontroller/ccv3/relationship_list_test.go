package ccv3_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("RelationshipList", func() {
	var (
		client *Client
	)

	BeforeEach(func() {
		client = NewTestClient()
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

	Describe("ShareServiceInstanceToSpaces", func() {
		var (
			serviceInstanceGUID string
			spaceGUIDs          []string

			relationshipList RelationshipList
			warnings         Warnings
			executeErr       error
		)

		BeforeEach(func() {
			serviceInstanceGUID = "some-service-instance-guid"
			spaceGUIDs = []string{"some-space-guid", "some-other-space-guid"}
		})

		JustBeforeEach(func() {
			relationshipList, warnings, executeErr = client.ShareServiceInstanceToSpaces(serviceInstanceGUID, spaceGUIDs)
		})

		Context("when no errors are encountered", func() {
			BeforeEach(func() {
				response := `{
					"data": [
						{
							"guid": "some-space-guid"
						},
						{
							"guid": "some-other-space-guid"
						}
					]
				}`

				requestBody := map[string][]map[string]string{
					"data": {{"guid": "some-space-guid"}, {"guid": "some-other-space-guid"}},
				}

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/service_instances/some-service-instance-guid/relationships/shared_spaces"),
						VerifyJSONRepresenting(requestBody),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns all relationships and warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))
				Expect(relationshipList).To(Equal(RelationshipList{
					GUIDs: []string{"some-space-guid", "some-other-space-guid"},
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
						VerifyRequest(http.MethodPost, "/v3/service_instances/some-service-instance-guid/relationships/shared_spaces"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and all warnings", func() {
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
})
