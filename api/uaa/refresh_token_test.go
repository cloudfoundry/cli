package uaa_test

import (
	"fmt"
	"net/http"

	. "code.cloudfoundry.org/cli/api/uaa"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("UAA Client", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestUAAClientAndStore()
	})

	Describe("RefreshAccessToken", func() {
		var (
			returnedAccessToken  string
			sentRefreshToken     string
			returnedRefreshToken string
		)

		BeforeEach(func() {
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
					VerifyBody([]byte(fmt.Sprintf("client_id=client-id&client_secret=client-secret&grant_type=refresh_token&refresh_token=%s", sentRefreshToken))),
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
