package sharedaction_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/sharedaction/sharedactionfakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AuthActor", func() {
	var (
		actor      *Actor
		fakeConfig *sharedactionfakes.FakeConfig
	)

	BeforeEach(func() {
		fakeConfig = new(sharedactionfakes.FakeConfig)
	})

	Context("Default CF on VMs", func() {
		BeforeEach(func() {
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

	Context("CF on K8s", func() {
		BeforeEach(func() {
			fakeConfig.IsCFOnK8sReturns(true)
			actor = NewActor(fakeConfig)
		})

		When("the auth info is set", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserNameReturns("non-empty", nil)
			})

			It("returns true", func() {
				Expect(actor.IsLoggedIn()).To(BeTrue())
			})
		})

		When("the auth info is not set", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserNameReturns("", nil)
			})

			It("returns false", func() {
				Expect(actor.IsLoggedIn()).To(BeFalse())
			})
		})

		When("getting the current user name fails", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserNameReturns("", errors.New("boom!"))
			})

			It("returns false", func() {
				Expect(actor.IsLoggedIn()).To(BeFalse())
			})
		})
	})
})
