package v2action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/uaa"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Token Actions", func() {
	var (
		actor         *Actor
		fakeConfig    *v2actionfakes.FakeConfig
		fakeUAAClient *v2actionfakes.FakeUAAClient
	)

	BeforeEach(func() {
		fakeConfig = new(v2actionfakes.FakeConfig)
		fakeUAAClient = new(v2actionfakes.FakeUAAClient)
		actor = NewActor(nil, fakeUAAClient, fakeConfig)
	})

	Describe("RefreshAccessToken", func() {
		Context("when an error is encountered refreshing the access token", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("refresh tokens error")
				fakeUAAClient.RefreshAccessTokenReturns(uaa.RefreshedTokens{}, expectedErr)
			})

			It("does not save any tokens to config and returns the error", func() {
				_, err := actor.RefreshAccessToken("existing-refresh-token")
				Expect(err).To(MatchError(expectedErr))

				Expect(fakeUAAClient.RefreshAccessTokenCallCount()).To(Equal(1))
				Expect(fakeUAAClient.RefreshAccessTokenArgsForCall(0)).To(Equal("existing-refresh-token"))

				Expect(fakeConfig.SetRefreshTokenCallCount()).To(Equal(0))
			})
		})

		Context("when no errors are encountered refreshing the access token", func() {
			BeforeEach(func() {
				fakeUAAClient.RefreshAccessTokenReturns(
					uaa.RefreshedTokens{
						AccessToken:  "new-access-token",
						RefreshToken: "new-refresh-token",
						Type:         "bob",
					},
					nil)
			})

			It("saves the new access and refresh tokens in the config and returns the access token", func() {
				accessToken, err := actor.RefreshAccessToken("existing-refresh-token")
				Expect(err).ToNot(HaveOccurred())
				Expect(accessToken).To(Equal("bob new-access-token"))

				Expect(fakeUAAClient.RefreshAccessTokenCallCount()).To(Equal(1))
				Expect(fakeUAAClient.RefreshAccessTokenArgsForCall(0)).To(Equal("existing-refresh-token"))

				Expect(fakeConfig.SetAccessTokenCallCount()).To(Equal(1))
				Expect(fakeConfig.SetAccessTokenArgsForCall(0)).To(Equal("bob new-access-token"))

				Expect(fakeConfig.SetRefreshTokenCallCount()).To(Equal(1))
				Expect(fakeConfig.SetRefreshTokenArgsForCall(0)).To(Equal("new-refresh-token"))
			})
		})
	})
})
