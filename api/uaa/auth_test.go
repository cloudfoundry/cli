package uaa_test

import (
	"fmt"
	"net/http"

	. "code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Auth", func() {
	var (
		client *Client
	)

	BeforeEach(func() {
		client = NewTestUAAClientAndStore()
	})

	Describe("Authenticate", func() {
		Context("when no errors occur", func() {
			var (
				username string
				password string
			)

			BeforeEach(func() {
				response := `{
						"access_token":"some-access-token",
						"refresh_token":"some-refresh-token"
					}`
				username = helpers.NewUsername()
				password = helpers.NewPassword()
				server.AppendHandlers(
					CombineHandlers(
						verifyRequestHost(TestAuthorizationResource),
						VerifyRequest(http.MethodPost, "/oauth/token"),
						VerifyHeaderKV("Content-Type", "application/x-www-form-urlencoded"),
						VerifyHeaderKV("Authorization", "Basic Y2xpZW50LWlkOmNsaWVudC1zZWNyZXQ="),
						VerifyBody([]byte(fmt.Sprintf("grant_type=password&password=%s&username=%s", password, username))),
						RespondWith(http.StatusOK, response),
					))
			})

			It("authenticates with the credentials provided", func() {
				accessToken, refreshToken, err := client.Authenticate(username, password)
				Expect(err).NotTo(HaveOccurred())

				Expect(accessToken).To(Equal("some-access-token"))
				Expect(refreshToken).To(Equal("some-refresh-token"))
			})
		})

		Context("when an error occurs", func() {
			var response string

			BeforeEach(func() {
				response = `{
						"error": "some-error",
						"error_description": "some-description"
					}`
				server.AppendHandlers(
					CombineHandlers(
						verifyRequestHost(TestAuthorizationResource),
						VerifyRequest(http.MethodPost, "/oauth/token"),
						RespondWith(http.StatusTeapot, response),
					))
			})

			It("returns the error", func() {
				_, _, err := client.Authenticate("us3r", "pa55")
				Expect(err).To(MatchError(RawHTTPStatusError{
					StatusCode:  http.StatusTeapot,
					RawResponse: []byte(response),
				}))
			})
		})
	})
})
