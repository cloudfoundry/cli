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
		binaryName string
		fakeConfig *sharedactionfakes.FakeConfig
	)

	BeforeEach(func() {
		binaryName = "faceman"
		fakeConfig = new(sharedactionfakes.FakeConfig)
		fakeConfig.BinaryNameReturns(binaryName)
		actor = NewActor(fakeConfig)
	})

	When("the access token is not set", func() {
		BeforeEach(func() {
			fakeConfig.AccessTokenReturns("")
		})

		It("returns false", func() {
			Expect(actor.IsLoggedIn()).To(BeFalse())
		})
	})

	When("the refresh token is not set", func() {
		BeforeEach(func() {
			fakeConfig.RefreshTokenReturns("")
		})

		It("returns false", func() {
			Expect(actor.IsLoggedIn()).To(BeFalse())
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
})
