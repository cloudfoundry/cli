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

var _ = Describe("UAA Client", func() {
	var (
		client *Client

		fakeConfig *uaafakes.FakeConfig
	)

	BeforeEach(func() {
		fakeConfig = NewTestConfig()

		client = NewTestUAAClientAndStore(fakeConfig)
	})

	Describe("RefreshAccessToken", func() {
		var (
			returnedAccessToken  string
			sentRefreshToken     string
			returnedRefreshToken string
		)

		When("the provided grant_type is client_credentials", func() {
			BeforeEach(func() {
				fakeConfig.UAAGrantTypeReturns(string(constant.GrantTypeClientCredentials))

				returnedAccessToken = "I-ACCESS-TOKEN"
				response := fmt.Sprintf(`{
				"access_token": "%s",
				"token_type": "bearer",
				"expires_in": 599,
				"scope": "cloud_controller.read password.write cloud_controller.write openid uaa.user",
				"jti": "4150c08afa2848278e5ad57201024e32"
			}`, returnedAccessToken)

				server.AppendHandlers(
					CombineHandlers(
						verifyRequestHost(TestAuthorizationResource),
						VerifyRequest(http.MethodPost, "/oauth/token"),
						VerifyHeaderKV("Accept", "application/json"),
						VerifyHeaderKV("Content-Type", "application/x-www-form-urlencoded"),
						VerifyHeaderKV("Authorization"),
						VerifyBody([]byte(fmt.Sprintf("client_id=client-id&client_secret=client-secret&grant_type=%s", constant.GrantTypeClientCredentials))),
						RespondWith(http.StatusOK, response),
					))
			})

			It("refreshes the tokens", func() {
				token, err := client.RefreshAccessToken(sentRefreshToken)
				Expect(err).ToNot(HaveOccurred())
				Expect(token).To(Equal(RefreshedTokens{
					AccessToken: returnedAccessToken,
					Type:        "bearer",
				}))

				Expect(server.ReceivedRequests()).To(HaveLen(2))
			})
		})

		When("the provided grant_type is password", func() {
			BeforeEach(func() {
				fakeConfig.UAAGrantTypeReturns(string(constant.GrantTypePassword))

				returnedAccessToken = "I-ACCESS-TOKEN"
				sentRefreshToken = "I-R-REFRESH-TOKEN"
				returnedRefreshToken = "I-R-NEW-REFRESH-TOKEN"
				response := fmt.Sprintf(`{
				"access_token": "%s",
				"token_type": "bearer",
				"refresh_token": "%s",
				"expires_in": 599,
				"scope": "cloud_controller.read password.write cloud_controller.write openid uaa.user",
				"jti": "4150c08afa2848278e5ad57201024e32"
			}`, returnedAccessToken, returnedRefreshToken)

				server.AppendHandlers(
					CombineHandlers(
						verifyRequestHost(TestAuthorizationResource),
						VerifyRequest(http.MethodPost, "/oauth/token"),
						VerifyHeaderKV("Accept", "application/json"),
						VerifyHeaderKV("Content-Type", "application/x-www-form-urlencoded"),
						VerifyHeaderKV("Authorization", "Basic Y2xpZW50LWlkOmNsaWVudC1zZWNyZXQ="),
						VerifyBody([]byte(fmt.Sprintf("grant_type=%s&refresh_token=%s", constant.GrantTypeRefreshToken, sentRefreshToken))),
						RespondWith(http.StatusOK, response),
					))
			})

			It("refreshes the tokens", func() {
				token, err := client.RefreshAccessToken(sentRefreshToken)
				Expect(err).ToNot(HaveOccurred())
				Expect(token).To(Equal(RefreshedTokens{
					AccessToken:  returnedAccessToken,
					RefreshToken: returnedRefreshToken,
					Type:         "bearer",
				}))

				Expect(server.ReceivedRequests()).To(HaveLen(2))
			})
		})

		When("the provided grant_type is empty", func() {
			BeforeEach(func() {
				fakeConfig.UAAGrantTypeReturns("")

				returnedAccessToken = "I-ACCESS-TOKEN"
				sentRefreshToken = "I-R-REFRESH-TOKEN"
				returnedRefreshToken = "I-R-NEW-REFRESH-TOKEN"
				response := fmt.Sprintf(`{
				"access_token": "%s",
				"token_type": "bearer",
				"refresh_token": "%s",
				"expires_in": 599,
				"scope": "cloud_controller.read password.write cloud_controller.write openid uaa.user",
				"jti": "4150c08afa2848278e5ad57201024e32"
			}`, returnedAccessToken, returnedRefreshToken)

				server.AppendHandlers(
					CombineHandlers(
						verifyRequestHost(TestAuthorizationResource),
						VerifyRequest(http.MethodPost, "/oauth/token"),
						VerifyHeaderKV("Accept", "application/json"),
						VerifyHeaderKV("Content-Type", "application/x-www-form-urlencoded"),
						VerifyHeaderKV("Authorization", "Basic Y2xpZW50LWlkOmNsaWVudC1zZWNyZXQ="),
						VerifyBody([]byte(fmt.Sprintf("grant_type=%s&refresh_token=%s", constant.GrantTypeRefreshToken, sentRefreshToken))),
						RespondWith(http.StatusOK, response),
					))
			})

			It("refreshes the tokens", func() {
				token, err := client.RefreshAccessToken(sentRefreshToken)
				Expect(err).ToNot(HaveOccurred())
				Expect(token).To(Equal(RefreshedTokens{
					AccessToken:  returnedAccessToken,
					RefreshToken: returnedRefreshToken,
					Type:         "bearer",
				}))

				Expect(server.ReceivedRequests()).To(HaveLen(2))
			})
		})
	})
})
