package ccv3_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("User", func() {
	var client *Client

	BeforeEach(func() {
		client, _ = NewTestClient()
	})

	Describe("CreateUser", func() {
		var (
			warnings   Warnings
			executeErr error
			user       User
		)

		JustBeforeEach(func() {
			user, warnings, executeErr = client.CreateUser("some-uaa-guid")
		})

		When("no error occurs", func() {
			BeforeEach(func() {
				response := `{
				"guid": "some-uaa-guid",
				"username": "some-user-name",
                "presentation_name": "some-user-name",
                "origin": "ldap",
				"created_at": "2016-12-07T18:18:30Z",
				"updated_at": null

				}`
				expectedBody := `{"guid":"some-uaa-guid"}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/users"),
						VerifyJSON(expectedBody),
						RespondWith(http.StatusCreated, response, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					),
				)
			})

			It("creates and returns the user and all warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(user).To(Equal(User{GUID: "some-uaa-guid", Username: "some-user-name", PresentationName: "some-user-name", Origin: "ldap"}))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
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
    },
		{
      "code": 10010,
      "detail": "Isolation segment not found",
      "title": "CF-ResourceNotFound"
    }
  ]
}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/users"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and all warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.MultiError{
					ResponseCode: http.StatusTeapot,
					Errors: []ccerror.V3Error{
						{
							Code:   10008,
							Detail: "The request is semantically invalid: command presence",
							Title:  "CF-UnprocessableEntity",
						},
						{
							Code:   10010,
							Detail: "Isolation segment not found",
							Title:  "CF-ResourceNotFound",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("DeleteUser", func() {
		var (
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			warnings, executeErr = client.DeleteUser("some-uaa-guid")
		})

		When("no error occurs", func() {
			BeforeEach(func() {
				response := `{}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodDelete, "/v3/users/some-uaa-guid"),
						RespondWith(http.StatusCreated, response, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					),
				)
			})

			It("deletes and returns all warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})

		When("the cloud controller returns errors and warnings", func() {
			When("the error should be raise", func() {

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
      "detail": "Isolation segment not found",
      "title": "CF-ResourceNotFound"
    }
  ]
}`
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodDelete, "/v3/users/some-uaa-guid"),
							RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a delete warning"}}),
						),
					)
				})

				It("returns the error and all warnings", func() {
					Expect(executeErr).To(MatchError(ccerror.MultiError{
						ResponseCode: http.StatusTeapot,
						Errors: []ccerror.V3Error{
							{
								Code:   10008,
								Detail: "The request is semantically invalid: command presence",
								Title:  "CF-UnprocessableEntity",
							},
							{
								Code:   10010,
								Detail: "Isolation segment not found",
								Title:  "CF-ResourceNotFound",
							},
						},
					}))
					Expect(warnings).To(ConsistOf("this is a delete warning"))
				})
			})
		})
	})
})
