package ccv3_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("RelationshipList", func() {
	var (
		client *Client
	)

	BeforeEach(func() {
		client, _ = NewTestClient()
	})

	Describe("EntitleIsolationSegmentToOrganizations", func() {
		When("the delete is successful", func() {
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
				Expect(relationships).To(Equal(resources.RelationshipList{
					GUIDs: []string{"some-relationship-guid-1", "some-relationship-guid-2"},
				}))
			})
		})

		When("the cloud controller returns errors and warnings", func() {
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
})
