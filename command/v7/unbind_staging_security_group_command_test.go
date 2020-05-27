package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("unbind-staging-security-group Command", func() {
	var (
		cmd             UnbindStagingSecurityGroupCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		cmd = UnbindStagingSecurityGroupCommand{
			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)

		cmd.RequiredArgs.SecurityGroup = "some-security-group"

		fakeConfig.CurrentUserReturns(
			configv3.User{Name: "some-user"},
			nil)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("the command executes successfully", func() {
		BeforeEach(func() {
			fakeActor.UpdateSecurityGroupGloballyEnabledReturns(
				v7action.Warnings{"get security group warning"},
				nil)
		})

		It("unbinds the security group from being available globally for staging", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(testUI.Out).To(Say("Unbinding security group %s from defaults for staging as some-user...", cmd.RequiredArgs.SecurityGroup))
			Expect(testUI.Out).To(Say("OK"))
			Expect(testUI.Out).To(Say(`TIP: Changes require an app restart \(for running\) or restage \(for staging\) to apply to existing applications\.`))

			Expect(testUI.Err).To(Say("get security group warning"))
		})
	})

	When("the security group does not exist", func() {
		BeforeEach(func() {
			fakeActor.UpdateSecurityGroupGloballyEnabledReturns(
				v7action.Warnings{"update security group warning"},
				actionerror.SecurityGroupNotFoundError{Name: cmd.RequiredArgs.SecurityGroup})
		})

		It("displays OK, warning and exits without error", func() {
			Expect(testUI.Err).To(Say("update security group warning"))
			Expect(testUI.Err).To(Say("Security group '%s' not found.", cmd.RequiredArgs.SecurityGroup))
			Expect(testUI.Out).To(Say("OK"))
			Expect(executeErr).NotTo(HaveOccurred())
		})
	})

	When("updating the security group fails", func() {
		BeforeEach(func() {
			fakeActor.UpdateSecurityGroupGloballyEnabledReturns(
				v7action.Warnings{"update security group warning"},
				errors.New("update security group error"))
		})

		It("returns the error", func() {
			Expect(testUI.Err).To(Say("update security group warning"))
			Expect(executeErr).To(MatchError("update security group error"))
		})
	})
})
