package uaa_test

import (
	"fmt"
	"net/http"

	. "code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/api/uaa/constant"
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
		var (
			identity  string
			secret    string
			grantType constant.GrantType

			accessToken  string
			refreshToken string
			executeErr   error
		)

		JustBeforeEach(func() {
			accessToken, refreshToken, executeErr = client.Authenticate(identity, secret, grantType)
		})

		Context("when no errors occur", func() {
			Context("when the grant type is password", func() {
				BeforeEach(func() {
					response := `{
						"access_token":"some-access-token",
						"refresh_token":"some-refresh-token"
					}`
					identity = helpers.NewUsername()
					secret = helpers.NewPassword()
					grantType = constant.GrantTypePassword
					server.AppendHandlers(
						CombineHandlers(
							verifyRequestHost(TestAuthorizationResource),
							VerifyRequest(http.MethodPost, "/oauth/token"),
							VerifyHeaderKV("Content-Type", "application/x-www-form-urlencoded"),
							VerifyHeaderKV("Authorization", "Basic Y2xpZW50LWlkOmNsaWVudC1zZWNyZXQ="),
							VerifyBody([]byte(fmt.Sprintf("grant_type=%s&password=%s&username=%s", grantType, secret, identity))),
							RespondWith(http.StatusOK, response),
						))
				})

				It("authenticates with the credentials provided", func() {
					Expect(executeErr).NotTo(HaveOccurred())

					Expect(accessToken).To(Equal("some-access-token"))
					Expect(refreshToken).To(Equal("some-refresh-token"))
				})
			})

			Context("when the grant type is client credentials", func() {
				BeforeEach(func() {
					response := `{
						"access_token":"some-access-token"
					}`

					identity = helpers.NewUsername()
					secret = helpers.NewPassword()
					grantType = constant.GrantTypeClientCredentials
					server.AppendHandlers(
						CombineHandlers(
							verifyRequestHost(TestAuthorizationResource),
							VerifyRequest(http.MethodPost, "/oauth/token"),
							VerifyHeaderKV("Content-Type", "application/x-www-form-urlencoded"),
							VerifyHeaderKV("Authorization", "Basic Y2xpZW50LWlkOmNsaWVudC1zZWNyZXQ="),
							VerifyBody([]byte(fmt.Sprintf("client_id=%s&client_secret=%s&grant_type=%s", identity, secret, grantType))),
							RespondWith(http.StatusOK, response),
						))
				})

				It("authenticates with the credentials provided", func() {
					Expect(executeErr).NotTo(HaveOccurred())

					Expect(accessToken).To(Equal("some-access-token"))
					Expect(refreshToken).To(BeEmpty())
				})
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
				Expect(executeErr).To(MatchError(RawHTTPStatusError{
					StatusCode:  http.StatusTeapot,
					RawResponse: []byte(response),
				}))
			})
		})
	})
})
