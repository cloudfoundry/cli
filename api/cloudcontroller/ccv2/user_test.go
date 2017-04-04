package ccv2_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("User", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("CreateUser", func() {
		Context("when an error does not occur", func() {
			BeforeEach(func() {
				response := `{
					 "metadata": {
							"guid": "some-guid",
							"url": "/v2/users/some-guid",
							"created_at": "2016-12-07T18:18:30Z",
							"updated_at": null
					 },
					 "entity": {
							"admin": false,
							"active": false,
							"default_space_guid": null,
							"spaces_url": "/v2/users/some-guid/spaces",
							"organizations_url": "/v2/users/some-guid/organizations",
							"managed_organizations_url": "/v2/users/some-guid/managed_organizations",
							"billing_managed_organizations_url": "/v2/users/some-guid/billing_managed_organizations",
							"audited_organizations_url": "/v2/users/some-guid/audited_organizations",
							"managed_spaces_url": "/v2/users/some-guid/managed_spaces",
							"audited_spaces_url": "/v2/users/some-guid/audited_spaces"
					 }
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v2/users"),
						VerifyJSON(`{"guid":"some-uaa-guid"}`),
						RespondWith(http.StatusCreated, response, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					),
				)
			})

			It("creates and returns the user and all warnings", func() {
				user, warnings, err := client.CreateUser("some-uaa-guid")
				Expect(err).ToNot(HaveOccurred())

				Expect(user).To(Equal(User{GUID: "some-guid"}))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})

		Context("when cloud controller returns an error and warnings", func() {
			BeforeEach(func() {
				response := `{
					"code": 10008,
					"description": "The request is semantically invalid: command presence",
					"error_code": "CF-UnprocessableEntity"
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v2/users"),
						VerifyJSON(`{"guid":"some-uaa-guid"}`),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					))
			})

			It("returns the errors and all warnings", func() {
				_, warnings, err := client.CreateUser("some-uaa-guid")
				Expect(err).To(MatchError(ccerror.V2UnexpectedResponseError{
					ResponseCode: 418,
					V2ErrorResponse: ccerror.V2ErrorResponse{
						Code:        10008,
						Description: "The request is semantically invalid: command presence",
						ErrorCode:   "CF-UnprocessableEntity",
					},
				}))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})
	})
})
