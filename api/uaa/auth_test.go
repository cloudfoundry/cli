package uaa_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	. "code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/api/uaa/constant"
	"code.cloudfoundry.org/cli/api/uaa/uaafakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Auth", func() {
	var (
		client      *Client
		credentials map[string]string

		origin    string
		grantType constant.GrantType

		accessToken  string
		refreshToken string
		executeErr   error
		fakeConfig   *uaafakes.FakeConfig
	)

	BeforeEach(func() {
		fakeConfig = NewTestConfig()

		client = NewTestUAAClientAndStore(fakeConfig)
	})

	Describe("Authenticate", func() {
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

	Describe("Revoke", func() {
		var (
			jtiFromToken string
			testToken    string
			actualError  error
		)

		JustBeforeEach(func() {
			actualError = client.Revoke(testToken)
		})

		When("the call to revoke succeeds", func() {
			BeforeEach(func() {
				testToken = "eyJhbGciOiJIUzI1NiJ9.eyJqdGkiOiI1NWZiOTVlM2M5OTY0MmYxODQxMTUyZWIwNmFjYTM4NiIsInJldm9jYWJsZSI6dHJ1ZX0.EYvzQDsCqXRO0r1dowPRYIgeqP_V320v1WbLTG5y6iA"
				jtiFromToken = "55fb95e3c99642f1841152eb06aca386"
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodDelete, fmt.Sprintf("/oauth/token/revoke/%s", jtiFromToken)),
						VerifyHeaderKV("Authorization: Bearer "+testToken),
						RespondWith(http.StatusOK, nil),
					))
			})

			It("makes a call to find the jti and uses that jti to revoke the token", func() {
				Expect(actualError).To(BeNil())
				Expect(len(server.ReceivedRequests())).To(Equal(1))
				Expect(server.ReceivedRequests()[0].RequestURI).To(Equal("/oauth/token/revoke/55fb95e3c99642f1841152eb06aca386"))
			})
		})

		When("the call to revoke the token fails", func() {
			BeforeEach(func() {
				testToken = "eyJhbGciOiJIUzI1NiJ9.eyJqdGkiOiI1NWZiOTVlM2M5OTY0MmYxODQxMTUyZWIwNmFjYTM4NiIsInJldm9jYWJsZSI6dHJ1ZX0.EYvzQDsCqXRO0r1dowPRYIgeqP_V320v1WbLTG5y6iA"
				jtiFromToken = "55fb95e3c99642f1841152eb06aca386"

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodDelete, fmt.Sprintf("/oauth/token/revoke/%s", jtiFromToken)),
						VerifyHeaderKV("Authorization: Bearer "+testToken),
						RespondWith(http.StatusForbidden, nil),
					))

			})
			It("returns that failure", func() {
				Expect(actualError).To(MatchError(RawHTTPStatusError{StatusCode: 403, RawResponse: []byte{}}))
			})
		})

		When("the token contains no jti", func() {
			BeforeEach(func() {
				testToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"
			})
			It("returns an error saying it could not parse the jti", func() {
				Expect(actualError).To(Equal(errors.New("could not parse jti from payload")))
			})
		})
	})
})
