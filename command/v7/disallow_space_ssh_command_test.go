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

var _ = Describe("disallow-space-ssh Command", func() {
	var (
		cmd                   DisallowSpaceSSHCommand
		testUI                *ui.UI
		fakeConfig            *commandfakes.FakeConfig
		fakeSharedActor       *commandfakes.FakeSharedActor
		disallowSpaceSSHActor *v7fakes.FakeDisallowSpaceSSHActor

		binaryName      string
		currentUserName string
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		disallowSpaceSSHActor = new(v7fakes.FakeDisallowSpaceSSHActor)

		cmd = DisallowSpaceSSHCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       disallowSpaceSSHActor,
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
				disallowSpaceSSHActor.UpdateSpaceFeatureReturns(
					v7action.Warnings{"some-warning"},
					nil,
				)
			})

			It("disallows ssh for the space", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(disallowSpaceSSHActor.UpdateSpaceFeatureCallCount()).To(Equal(1))

				Expect(testUI.Out).To(Say("Disabling ssh support for space %s as %s...", cmd.RequiredArgs.Space, currentUserName))
				Expect(testUI.Out).To(Say("OK"))
				Expect(testUI.Err).To(Say("some-warning"))
			})
		})

		When("ssh is already disallowed", func() {
			BeforeEach(func() {
				disallowSpaceSSHActor.UpdateSpaceFeatureReturns(
					v7action.Warnings{"some-warning"},
					actionerror.SpaceSSHAlreadyDisabledError{Space: "some-space"},
				)
			})

			It("allows ssh for the space", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(disallowSpaceSSHActor.UpdateSpaceFeatureCallCount()).To(Equal(1))

				Expect(testUI.Out).To(Say("Disabling ssh support for space %s as %s...", cmd.RequiredArgs.Space, currentUserName))
				Expect(testUI.Out).To(Say("ssh support for space '%s' is already disabled.", cmd.RequiredArgs.Space))
				Expect(testUI.Out).To(Say("OK"))
				Expect(testUI.Err).To(Say("some-warning"))
			})
		})

		When("an error occurs while disabling SSH", func() {
			BeforeEach(func() {
				disallowSpaceSSHActor.UpdateSpaceFeatureReturns(
					v7action.Warnings{"some-warning"},
					errors.New("disallow-ssh-error"),
				)
			})

			It("does not display OK and returns the error", func() {
				Expect(testUI.Out).NotTo(Say("OK"))
				Expect(testUI.Err).To(Say("some-warning"))
				Expect(executeErr).To(MatchError("disallow-ssh-error"))
			})
		})
	})
})
