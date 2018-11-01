package sharedaction_test

import (
	. "code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/sharedaction/sharedactionfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("IsLoggedIn", func() {
	var (
		actor      *Actor
		fakeConfig *sharedactionfakes.FakeConfig
	)

	BeforeEach(func() {
		fakeConfig = new(sharedactionfakes.FakeConfig)
		actor = NewActor(fakeConfig)
	})

	When("only the access token is set", func() {
		BeforeEach(func() {
			fakeConfig.AccessTokenReturns("some-access-token")
		})

		It("returns true", func() {
			Expect(actor.IsLoggedIn()).To(BeTrue())
		})
	})

	When("only the refresh token is set", func() {
		BeforeEach(func() {
			fakeConfig.RefreshTokenReturns("some-refresh-token")
		})

		It("returns true", func() {
			Expect(actor.IsLoggedIn()).To(BeTrue())
		})
	})

	When("both access and refresh token are set", func() {
		BeforeEach(func() {
			fakeConfig.AccessTokenReturns("some-access-token")
			fakeConfig.RefreshTokenReturns("some-refresh-token")
		})

		It("returns true", func() {
			Expect(actor.IsLoggedIn()).To(BeTrue())
		})
	})

	When("neither access nor refresh token are set", func() {
		BeforeEach(func() {
			fakeConfig.AccessTokenReturns("")
			fakeConfig.RefreshTokenReturns("")
		})

		It("returns false", func() {
			Expect(actor.IsLoggedIn()).To(BeFalse())
		})
	})
})
