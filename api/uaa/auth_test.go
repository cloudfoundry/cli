package uaa_test

import (
	"fmt"
	"net/http"
	"net/url"

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
			credentials map[string]string

			origin    string
			grantType constant.GrantType

			accessToken  string
			refreshToken string
			executeErr   error
		)

		JustBeforeEach(func() {
			accessToken, refreshToken, executeErr = client.Authenticate(credentials, origin, grantType)
		})

		When("no errors occur", func() {
			When("the grant type is password", func() {
				var response string
				BeforeEach(func() {
					response = `{
						"access_token":"some-access-token",
						"refresh_token":"some-refresh-token"
					}`
					credentials = map[string]string{
						"username": "some-username",
						"password": "some-password",
					}
					grantType = constant.GrantTypePassword
				})

				When("origin is not set", func() {
					BeforeEach(func() {
						origin = ""
						server.AppendHandlers(
							CombineHandlers(
								verifyRequestHost(TestAuthorizationResource),
								VerifyRequest(http.MethodPost, "/oauth/token", ""),
								VerifyHeaderKV("Content-Type", "application/x-www-form-urlencoded"),
								VerifyHeaderKV("Authorization", "Basic Y2xpZW50LWlkOmNsaWVudC1zZWNyZXQ="),
								VerifyBody([]byte("grant_type=password&password=some-password&username=some-username")),
								RespondWith(http.StatusOK, response),
							))
					})

					It("authenticates with the credentials provided", func() {
						Expect(executeErr).NotTo(HaveOccurred())

						Expect(accessToken).To(Equal("some-access-token"))
						Expect(refreshToken).To(Equal("some-refresh-token"))
					})
				})

				When("origin is set", func() {
					BeforeEach(func() {
						origin = "some-fake-origin"
						expectedQuery := "login_hint=%7B%22origin%22%3A%22" + origin + "%22%7D"
						server.AppendHandlers(
							CombineHandlers(
								verifyRequestHost(TestAuthorizationResource),
								VerifyRequest(http.MethodPost, "/oauth/token", expectedQuery),
								VerifyHeaderKV("Content-Type", "application/x-www-form-urlencoded"),
								VerifyHeaderKV("Authorization", "Basic Y2xpZW50LWlkOmNsaWVudC1zZWNyZXQ="),
								VerifyBody([]byte("grant_type=password&password=some-password&username=some-username")),
								RespondWith(http.StatusOK, response),
							))
					})

					It("authenticates with the credentials provided", func() {
						Expect(executeErr).NotTo(HaveOccurred())

						Expect(accessToken).To(Equal("some-access-token"))
						Expect(refreshToken).To(Equal("some-refresh-token"))
					})
				})

				When("additional prompts are answered", func() {
					BeforeEach(func() {
						credentials = map[string]string{
							"username":     "some-username",
							"password":     "some-password",
							"mfaCode":      "some-mfa-code",
							"customPrompt": "some-custom-value",
						}

						expectedValues := url.Values{
							"username":     []string{"some-username"},
							"password":     []string{"some-password"},
							"mfaCode":      []string{"some-mfa-code"},
							"customPrompt": []string{"some-custom-value"},
						}

						server.AppendHandlers(
							CombineHandlers(
								VerifyForm(expectedValues),
								RespondWith(http.StatusOK, response),
							),
						)
					})

					It("sends all the prompts to the UAA", func() {
						Expect(executeErr).NotTo(HaveOccurred())
						Expect(accessToken).To(Equal("some-access-token"))
						Expect(refreshToken).To(Equal("some-refresh-token"))
					})
				})
			})

			When("the grant type is client credentials", func() {
				BeforeEach(func() {
					response := `{
						"access_token":"some-access-token"
					}`

					credentials = map[string]string{
						"client_id":     "some-client-id",
						"client_secret": "some-client-secret",
					}
					origin = ""
					grantType = constant.GrantTypeClientCredentials
					server.AppendHandlers(
						CombineHandlers(
							verifyRequestHost(TestAuthorizationResource),
							VerifyRequest(http.MethodPost, "/oauth/token"),
							VerifyHeaderKV("Content-Type", "application/x-www-form-urlencoded"),
							VerifyHeaderKV("Authorization"),
							VerifyBody([]byte(fmt.Sprintf("client_id=%s&client_secret=%s&grant_type=%s", credentials["client_id"], credentials["client_secret"], grantType))),
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
