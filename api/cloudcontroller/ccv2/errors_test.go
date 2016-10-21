package ccv2_test

import (
	"net/http"

	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Cloud Controller Connection", func() {
	var (
		response           string
		serverResponseCode int

		client *CloudControllerClient
	)

	BeforeEach(func() {
		response = `{
					"code": 777,
					"description": "SomeCC Error Message",
					"error_code": "CF-SomeError"
				}`

		client = NewTestClient()
	})

	JustBeforeEach(func() {
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/v2/apps"),
				RespondWith(serverResponseCode, response),
			),
		)
	})

	Describe("Make", func() {
		Describe("(401) Unauthorized", func() {
			BeforeEach(func() {
				serverResponseCode = http.StatusUnauthorized
			})

			Context("generic 401", func() {
				It("returns a UnauthorizedError", func() {
					_, _, err := client.GetApplications(nil)
					Expect(err).To(MatchError(UnauthorizedError{Message: "SomeCC Error Message"}))
				})
			})

			Context("invalid token", func() {
				BeforeEach(func() {
					response = `{
						"code": 1000,
						"description": "Invalid Auth Token",
						"error_code": "CF-InvalidAuthToken"
					}`
				})

				It("returns a UnauthorizedError", func() {
					_, _, err := client.GetApplications(nil)
					Expect(err).To(MatchError(InvalidAuthTokenError{Message: "Invalid Auth Token"}))
				})
			})
		})

		Describe("(403) Forbidden", func() {
			BeforeEach(func() {
				serverResponseCode = http.StatusForbidden
			})

			It("returns a ForbiddenError", func() {
				_, _, err := client.GetApplications(nil)
				Expect(err).To(MatchError(ForbiddenError{Message: "SomeCC Error Message"}))
			})
		})

		Describe("(404) Not Found", func() {
			BeforeEach(func() {
				serverResponseCode = http.StatusNotFound
			})

			It("returns a ResourceNotFoundError", func() {
				_, _, err := client.GetApplications(nil)
				Expect(err).To(MatchError(ResourceNotFoundError{Message: "SomeCC Error Message"}))
			})
		})

		Describe("Unhandled Error Codes", func() {
			BeforeEach(func() {
				serverResponseCode = http.StatusTeapot
			})

			It("returns a ResourceNotFoundError", func() {
				_, _, err := client.GetApplications(nil)
				Expect(err).To(MatchError(UnexpectedResponseError{
					ResponseCode: http.StatusTeapot,
					CCErrorResponse: CCErrorResponse{
						Code:        777,
						Description: "SomeCC Error Message",
						ErrorCode:   "CF-SomeError",
					},
				}))
			})
		})
	})
})
