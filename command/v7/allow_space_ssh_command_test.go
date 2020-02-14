package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("allow-space-ssh Command", func() {
	var (
		cmd             AllowSpaceSSHCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor

		binaryName      string
		currentUserName string
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		cmd = AllowSpaceSSHCommand{
			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		cmd.RequiredArgs.Space = "some-space"

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		currentUserName = "some-user"
		fakeConfig.CurrentUserNameReturns(currentUserName, nil)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("checking target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NoOrganizationTargetedError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NoOrganizationTargetedError{BinaryName: binaryName}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeTrue())
			Expect(checkTargetedSpace).To(BeFalse())
		})
	})

	When("checking the current user fails", func() {
		BeforeEach(func() {
			fakeConfig.CurrentUserNameReturns("", errors.New("uh oh"))
		})

		It("returns the error", func() {
			Expect(executeErr).To(MatchError("uh oh"))
		})
	})

	When("the user is logged in", func() {
		When("no errors occur", func() {
			BeforeEach(func() {
				fakeActor.AllowSpaceSSHReturns(
					v7action.Warnings{"some-warning"},
					nil,
				)
			})

			It("allows ssh for the space", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeActor.AllowSpaceSSHCallCount()).To(Equal(1))

				Expect(testUI.Out).To(Say("Enabling ssh support for space %s as %s...", cmd.RequiredArgs.Space, currentUserName))
				Expect(testUI.Out).To(Say("OK"))
				Expect(testUI.Err).To(Say("some-warning"))
			})
		})

		When("ssh is already allowed", func() {
			BeforeEach(func() {
				fakeActor.AllowSpaceSSHReturns(
					v7action.Warnings{"some-warning"},
					actionerror.SpaceSSHAlreadyEnabledError{Space: "some-space"},
				)
			})

			It("allows ssh for the space", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeActor.AllowSpaceSSHCallCount()).To(Equal(1))

				Expect(testUI.Out).To(Say("Enabling ssh support for space %s as %s...", cmd.RequiredArgs.Space, currentUserName))
				Expect(testUI.Out).To(Say("ssh support for space '%s' is already enabled.", cmd.RequiredArgs.Space))
				Expect(testUI.Out).To(Say("OK"))
				Expect(testUI.Err).To(Say("some-warning"))
			})
		})

		When("an error occurs while enabling SSH", func() {
			BeforeEach(func() {
				fakeActor.AllowSpaceSSHReturns(
					v7action.Warnings{"some-warning"},
					errors.New("allow-ssh-error"),
				)
			})

			It("does not display OK and returns the error", func() {
				Expect(testUI.Out).NotTo(Say("OK"))
				Expect(testUI.Err).To(Say("some-warning"))
				Expect(executeErr).To(MatchError("allow-ssh-error"))
			})
		})
	})
})
