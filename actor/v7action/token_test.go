package v7action_test

import (
	"errors"
	"time"

	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/uaa"
	"github.com/SermoDigital/jose/crypto"
	"github.com/SermoDigital/jose/jws"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Token Actions", func() {
	var (
		actor         *Actor
		fakeUAAClient *v7actionfakes.FakeUAAClient
		fakeConfig    *v7actionfakes.FakeConfig
	)

	BeforeEach(func() {
		fakeUAAClient = new(v7actionfakes.FakeUAAClient)
		fakeConfig = new(v7actionfakes.FakeConfig)
		actor = NewActor(nil, fakeConfig, nil, fakeUAAClient, nil)
	})

	Describe("RefreshAccessToken", func() {
		var (
			accessToken string
			err         error
		)
		JustBeforeEach(func() {
			accessToken, err = actor.RefreshAccessToken()
		})

		When("the token is invalid", func() {
			BeforeEach(func() {
				fakeUAAClient.RefreshAccessTokenReturns(uaa.RefreshedTokens{
					AccessToken:  "some-token",
					Type:         "bearer",
					RefreshToken: "new-refresh-token",
				}, nil)
				fakeConfig.RefreshTokenReturns("some-refresh-token")
				fakeConfig.AccessTokenReturns("im a bad token :(")
			})

			It("returns the new access token from the uaa client", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(fakeUAAClient.RefreshAccessTokenCallCount()).To(Equal(1))
				Expect(fakeUAAClient.RefreshAccessTokenArgsForCall(0)).To(Equal("some-refresh-token"))
				Expect(accessToken).To(Equal("bearer some-token"))
			})

			It("updates the config with the refreshed token", func() {
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
				fakeConfig.RefreshTokenReturns("some-refresh-token")
				notExpiringAccessToken = buildTokenString(time.Now().AddDate(5, 0, 0))
				fakeConfig.AccessTokenReturns(notExpiringAccessToken)
			})

			It("returns the current token without refreshing it", func() {
				Expect(fakeUAAClient.RefreshAccessTokenCallCount()).To(Equal(0))
				Expect(err).ToNot(HaveOccurred())
				Expect(accessToken).To(Equal(notExpiringAccessToken))
			})
		})

		When("the token is about to expire", func() {
			BeforeEach(func() {
				fakeUAAClient.RefreshAccessTokenReturns(uaa.RefreshedTokens{
					AccessToken:  "some-token",
					RefreshToken: "new-refresh-token",
					Type:         "bearer",
				}, nil)
				fakeConfig.RefreshTokenReturns("some-refresh-token")
				expiringAccessToken := buildTokenString(time.Now().Add(5))
				fakeConfig.AccessTokenReturns(expiringAccessToken)
			})

			It("returns the new access token from the uaa client", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(fakeUAAClient.RefreshAccessTokenCallCount()).To(Equal(1))
				Expect(fakeUAAClient.RefreshAccessTokenArgsForCall(0)).To(Equal("some-refresh-token"))
				Expect(accessToken).To(Equal("bearer some-token"))
			})

			It("updates the config with the refreshed token", func() {
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

					Expect(accessToken).To(Equal(""))
					Expect(err).To(MatchError("I'm still an error!"))
				})
			})
		})
	})
})

func buildTokenString(expiration time.Time) string {
	c := jws.Claims{}
	c.SetExpiration(expiration)
	token := jws.NewJWT(c, crypto.Unsecured)
	tokenBytes, _ := token.Serialize(nil) //nolint: errcheck
	return string(tokenBytes)
}
