package noaabridge_test

import (
	"errors"

	"code.cloudfoundry.org/cli/api/uaa"
	. "code.cloudfoundry.org/cli/api/uaa/noaabridge"
	"code.cloudfoundry.org/cli/api/uaa/noaabridge/noaabridgefakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TokenRefresher", func() {
	Describe("RefreshAuthToken", func() {
		var (
			fakeUAAClient  *noaabridgefakes.FakeUAAClient
			fakeTokenCache *noaabridgefakes.FakeTokenCache
			tokenRefresher *TokenRefresher
		)

		BeforeEach(func() {
			fakeUAAClient = new(noaabridgefakes.FakeUAAClient)
			fakeTokenCache = new(noaabridgefakes.FakeTokenCache)
			tokenRefresher = NewTokenRefresher(fakeUAAClient, fakeTokenCache)
		})

		Context("when UAA communication is successful", func() {
			BeforeEach(func() {
				fakeTokenCache.RefreshTokenReturns("old-refresh-token")

				refreshToken := uaa.RefreshedTokens{
					AccessToken:  "some-access-token",
					RefreshToken: "some-refresh-token",
					Type:         "bearer",
				}
				fakeUAAClient.RefreshAccessTokenReturns(refreshToken, nil)
			})

			It("refreshes the token", func() {
				token, err := tokenRefresher.RefreshAuthToken()
				Expect(err).ToNot(HaveOccurred())
				Expect(token).To(Equal("bearer some-access-token"))

				Expect(fakeUAAClient.RefreshAccessTokenCallCount()).To(Equal(1))
				Expect(fakeUAAClient.RefreshAccessTokenArgsForCall(0)).To(Equal("old-refresh-token"))
			})

			It("stores the new access and refresh tokens", func() {
				_, err := tokenRefresher.RefreshAuthToken()
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeTokenCache.SetAccessTokenCallCount()).To(Equal(1))
				Expect(fakeTokenCache.SetAccessTokenArgsForCall(0)).To(Equal("bearer some-access-token"))
				Expect(fakeTokenCache.SetRefreshTokenCallCount()).To(Equal(1))
				Expect(fakeTokenCache.SetRefreshTokenArgsForCall(0)).To(Equal("some-refresh-token"))
			})
		})

		Context("when UAA communication returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("it's not working!!!!")
				fakeUAAClient.RefreshAccessTokenReturns(uaa.RefreshedTokens{}, expectedErr)
			})

			It("returns the error", func() {
				_, err := tokenRefresher.RefreshAuthToken()
				Expect(err).To(MatchError(expectedErr))
			})
		})
	})
})
