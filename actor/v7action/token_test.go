package v7action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/uaa"
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

	FDescribe("RefreshAccessToken", func() {

		BeforeEach(func() {
			fakeUAAClient.RefreshAccessTokenReturns(uaa.RefreshedTokens{
				AccessToken: "bearer some-token",
			}, nil)
			fakeConfig.RefreshTokenReturns("some-refresh-token")
		})
		It("returns the new access token from the uaa client", func() {
			accessToken, err := actor.RefreshAccessToken()
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeUAAClient.RefreshAccessTokenCallCount()).To(Equal(1))
			Expect(fakeUAAClient.RefreshAccessTokenArgsForCall(0)).To(Equal("some-refresh-token"))
			Expect(accessToken).To(Equal("bearer some-token"))

		})

		When("refreshing the access token fails", func() {
			BeforeEach(func() {
				fakeUAAClient.RefreshAccessTokenReturns(
					uaa.RefreshedTokens{},
					errors.New("I'm still an error!"),
				)
			})

			It("returns that error", func() {

				accessToken, err := actor.RefreshAccessToken()
				Expect(accessToken).To(Equal(""))
				Expect(err).To(MatchError("I'm still an error!"))
			})
		})
	})
})
