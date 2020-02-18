package v2action_test

import (
	"errors"
	"time"

	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/integration/helpers"
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
		var (
			accessToken string
			err         error
		)

		JustBeforeEach(func() {
			accessToken, err = actor.RefreshAccessToken("existing-refresh-token")
		})

		When("the access token is invalid", func() {
			BeforeEach(func() {
				fakeUAAClient.RefreshAccessTokenReturns(uaa.RefreshedTokens{
					AccessToken:  "some-token",
					Type:         "bearer",
					RefreshToken: "new-refresh-token",
				}, nil)

				fakeConfig.AccessTokenReturns("im a bad token :(")
			})

			It("returns the new access token from the uaa client", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(fakeUAAClient.RefreshAccessTokenCallCount()).To(Equal(1))
				Expect(fakeUAAClient.RefreshAccessTokenArgsForCall(0)).To(Equal("existing-refresh-token"))
				Expect(accessToken).To(Equal("bearer some-token"))
			})

			It("updates the config with the refreshed tokens", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(fakeConfig.SetAccessTokenCallCount()).To(Equal(1))
				Expect(fakeConfig.SetRefreshTokenCallCount()).To(Equal(1))
				Expect(fakeConfig.SetAccessTokenArgsForCall(0)).To(Equal("bearer some-token"))
				Expect(fakeConfig.SetRefreshTokenArgsForCall(0)).To(Equal("new-refresh-token"))
			})
		})

		When("the token is not about to expire", func() {
			var notExpiringAccessToken string
			BeforeEach(func() {
				fakeUAAClient.RefreshAccessTokenReturns(uaa.RefreshedTokens{
					AccessToken: "some-token",
					Type:        "bearer",
				}, nil)

				notExpiringAccessToken = helpers.BuildTokenString(time.Now().AddDate(5, 0, 0))
				fakeConfig.AccessTokenReturns(notExpiringAccessToken)
			})

			It("returns the current access token without refreshing it", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(fakeUAAClient.RefreshAccessTokenCallCount()).To(Equal(0))
				Expect(accessToken).To(Equal(notExpiringAccessToken))
			})

		})

		When("the access token is about to expire", func() {
			BeforeEach(func() {
				fakeUAAClient.RefreshAccessTokenReturns(uaa.RefreshedTokens{
					AccessToken:  "some-token",
					RefreshToken: "new-refresh-token",
					Type:         "bearer",
				}, nil)

				expiringAccessToken := helpers.BuildTokenString(time.Now().Add(5))
				fakeConfig.AccessTokenReturns(expiringAccessToken)
			})

			It("returns the new access token from the uaa client", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(fakeUAAClient.RefreshAccessTokenCallCount()).To(Equal(1))
				Expect(fakeUAAClient.RefreshAccessTokenArgsForCall(0)).To(Equal("existing-refresh-token"))
				Expect(accessToken).To(Equal("bearer some-token"))
			})

			It("updates the config with the refreshed tokens", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(fakeConfig.SetAccessTokenCallCount()).To(Equal(1))
				Expect(fakeConfig.SetRefreshTokenCallCount()).To(Equal(1))
				Expect(fakeConfig.SetAccessTokenArgsForCall(0)).To(Equal("bearer some-token"))
				Expect(fakeConfig.SetRefreshTokenArgsForCall(0)).To(Equal("new-refresh-token"))
			})

			When("refreshing the access token fails", func() {
				BeforeEach(func() {
					fakeUAAClient.RefreshAccessTokenReturns(
						uaa.RefreshedTokens{},
						errors.New("I'm still an error!"),
					)
				})

				It("returns that error", func() {
					Expect(err).To(MatchError("I'm still an error!"))
					Expect(accessToken).To(Equal(""), "AccessToken was not equal to \"\"")

					Expect(fakeUAAClient.RefreshAccessTokenCallCount()).To(Equal(1))
					Expect(fakeUAAClient.RefreshAccessTokenArgsForCall(0)).To(Equal("existing-refresh-token"))

					Expect(fakeConfig.SetAccessTokenCallCount()).To(Equal(0))
					Expect(fakeConfig.SetRefreshTokenCallCount()).To(Equal(0))
				})
			})
		})
	})
})
