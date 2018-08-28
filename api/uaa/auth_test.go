package uaa_test

import (
	"fmt"
	"net/http"

	. "code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/api/uaa/constant"
	"code.cloudfoundry.org/cli/api/uaa/uaafakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Auth", func() {
	var (
		client *Client

		fakeConfig *uaafakes.FakeConfig
	)

	BeforeEach(func() {
		fakeConfig = NewTestConfig()

		client = NewTestUAAClientAndStore(fakeConfig)
	})

	Describe("Authenticate", func() {
		var (
			identity  string
			secret    string
			origin    string
			grantType constant.GrantType

			accessToken  string
			refreshToken string
			executeErr   error
		)

		BeforeEach(func() {
			identity = "some-identity"
			secret = "some-secret"
		})

		JustBeforeEach(func() {
			accessToken, refreshToken, executeErr = client.Authenticate(identity, secret, origin, grantType)
		})

		When("no errors occur", func() {
			When("the grant type is password and origin is not set", func() {
				BeforeEach(func() {
					response := `{
						"access_token":"some-access-token",
						"refresh_token":"some-refresh-token"
					}`
					origin = ""
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

			When("the grant type is password and origin is set", func() {
				BeforeEach(func() {
					response := `{
						"access_token":"some-access-token",
						"refresh_token":"some-refresh-token"
					}`
					origin = "some-fake-origin"
					grantType = constant.GrantTypePassword
					expectedQuery := "login_hint=%7B%22origin%22%3A%22" + origin + "%22%7D"
					server.AppendHandlers(
						CombineHandlers(
							verifyRequestHost(TestAuthorizationResource),
							VerifyRequest(http.MethodPost, "/oauth/token", expectedQuery),
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

			When("the grant type is client credentials", func() {
				BeforeEach(func() {
					response := `{
						"access_token":"some-access-token"
					}`

					origin = ""
					grantType = constant.GrantTypeClientCredentials
					server.AppendHandlers(
						CombineHandlers(
							verifyRequestHost(TestAuthorizationResource),
							VerifyRequest(http.MethodPost, "/oauth/token"),
							VerifyHeaderKV("Content-Type", "application/x-www-form-urlencoded"),
							VerifyHeaderKV("Authorization"),
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

		When("an error occurs", func() {
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
