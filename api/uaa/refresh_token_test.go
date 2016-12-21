package uaa_test

import (
	"net/http"

	. "code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/api/uaa/uaafakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("UAA Client", func() {
	var (
		client    *Client
		fakeStore *uaafakes.FakeAuthenticationStore
	)

	BeforeEach(func() {
		client, fakeStore = NewTestUAAClientAndStore()
	})

	Describe("RefreshToken", func() {
		BeforeEach(func() {
			response := `{
				"access_token": "access-token",
				"token_type": "bearer",
				"refresh_token": "refresh-token",
				"expires_in": 599,
				"scope": "cloud_controller.read password.write cloud_controller.write openid uaa.user",
				"jti": "4150c08afa2848278e5ad57201024e32"
			}`
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/oauth/token"),
					VerifyHeaderKV("Accept", "application/json"),
					VerifyHeaderKV("Content-Type", "application/x-www-form-urlencoded"),
					VerifyBody([]byte("client_id=client-id&client_secret=client-secret&grant_type=refresh_token&refresh_token=refresh-token")),
					RespondWith(http.StatusOK, response),
				))
			fakeStore.RefreshTokenReturns("refresh-token")
			fakeStore.UAAOAuthClientReturns("client-id")
			fakeStore.UAAOAuthClientSecretReturns("client-secret")
		})

		It("refreshes the tokens", func() {
			err := client.RefreshToken()
			Expect(err).ToNot(HaveOccurred())

			Expect(server.ReceivedRequests()).To(HaveLen(1))

			Expect(fakeStore.SetAccessTokenCallCount()).To(Equal(1))
			Expect(fakeStore.SetAccessTokenArgsForCall(0)).To(Equal("bearer access-token"))

			Expect(fakeStore.SetRefreshTokenCallCount()).To(Equal(1))
			Expect(fakeStore.SetRefreshTokenArgsForCall(0)).To(Equal("refresh-token"))
		})
	})
})
