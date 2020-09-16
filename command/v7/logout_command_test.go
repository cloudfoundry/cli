package v7_test

import (
	"code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("logout command", func() {
	var (
		cmd        LogoutCommand
		testUI     *ui.UI
		fakeConfig *commandfakes.FakeConfig
		fakeActor  *v7fakes.FakeActor
		executeErr error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeActor = new(v7fakes.FakeActor)
		cmd = LogoutCommand{
			BaseCommand: BaseCommand{
				UI:     testUI,
				Config: fakeConfig,
				Actor:  fakeActor,
			},
		}

		fakeConfig.CurrentUserReturns(
			configv3.User{
				Name: "some-user",
			},
			nil)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	It("outputs logging out display message", func() {
		Expect(executeErr).ToNot(HaveOccurred())

		Expect(fakeConfig.UnsetUserInformationCallCount()).To(Equal(1))
		Expect(testUI.Out).To(Say("Logging out some-user..."))
		Expect(testUI.Out).To(Say("OK"))
	})

	It("calls to revoke the auth tokens", func() {
		Expect(fakeActor.RevokeAccessAndRefreshTokensCallCount()).To(Equal(1))
	})

	When("unable to revoke token", func() {
		When("because the user is not logged in", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserReturns(configv3.User{}, nil)
			})

			It("does not impact the logout", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeConfig.CurrentUserCallCount()).To(Equal(1))
				Expect(fakeConfig.UnsetUserInformationCallCount()).To(Equal(1))
				Expect(testUI.Out).To(Say("Logging out ..."))
				Expect(testUI.Out).To(Say("OK"))
			})
		})

		When("because the attempt to revoke fails", func() {
			BeforeEach(func() {
				fakeActor.RevokeAccessAndRefreshTokensReturns(error(uaa.UnauthorizedError{Message: "test error"}))
			})

			It("does not impact the logout", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeConfig.UnsetUserInformationCallCount()).To(Equal(1))
				Expect(testUI.Out).To(Say("Logging out some-user..."))
				Expect(testUI.Out).To(Say("OK"))
			})
		})
	})
})
