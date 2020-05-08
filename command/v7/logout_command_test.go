package v7_test

import (
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v7"
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
		executeErr error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		cmd = LogoutCommand{
			BaseCommand: command.BaseCommand{
				UI:     testUI,
				Config: fakeConfig,
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
})
