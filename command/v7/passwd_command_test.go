package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/translatableerror"
	v7 "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("passwd Command", func() {
	var (
		cmd             v7.PasswdCommand
		fakeUI          *commandfakes.FakeUI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		fakeUI = new(commandfakes.FakeUI)
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)

		cmd = v7.PasswdCommand{
			BaseCommand: command.BaseCommand{
				UI:          fakeUI,
				Config:      fakeConfig,
				Actor:       fakeActor,
				SharedActor: fakeSharedActor,
			},
		}

		fakeConfig.CurrentUserReturns(configv3.User{Name: "steve", GUID: "steve-guid"}, nil)

		fakeUI.DisplayPasswordPromptReturnsOnCall(0, "old1", nil)
		fakeUI.DisplayPasswordPromptReturnsOnCall(1, "new1", nil)
		fakeUI.DisplayPasswordPromptReturnsOnCall(2, "new1", nil)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("changing password succeeds", func() {
		It("displays flavor text", func() {
			text, variables := fakeUI.DisplayTextWithFlavorArgsForCall(0)
			Expect(text).To(Equal("Changing password for user {{.Username}}..."))
			Expect(variables[0]).To(Equal(map[string]interface{}{
				"Username": "steve",
			}))
		})

		It("makes the correct request to UAA", func() {
			Expect(fakeActor.UpdateUserPasswordCallCount()).To(Equal(1))

			userGUID, oldPassword, newPassword := fakeActor.UpdateUserPasswordArgsForCall(0)
			Expect(userGUID).To(Equal("steve-guid"))
			Expect(oldPassword).To(Equal("old1"))
			Expect(newPassword).To(Equal("new1"))
		})

		It("displays OK and unsets user information", func() {
			Expect(fakeUI.DisplayOKCallCount()).To(Equal(1))
			Expect(fakeConfig.UnsetUserInformationCallCount()).To(Equal(1))
			Expect(fakeUI.DisplayTextArgsForCall(0)).To(Equal("Please log in again."))
		})
	})

	When("checking target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NoOrganizationTargetedError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NoOrganizationTargetedError{BinaryName: binaryName}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeFalse())
			Expect(checkTargetedSpace).To(BeFalse())
		})
	})

	When("the user is not logged in", func() {
		var expectedErr error

		BeforeEach(func() {
			expectedErr = errors.New("some current user error")
			fakeConfig.CurrentUserReturns(configv3.User{}, expectedErr)
		})

		It("return an error", func() {
			Expect(executeErr).To(Equal(expectedErr))
		})
	})

	When("an error occurs while reading the current password", func() {
		BeforeEach(func() {
			fakeUI.DisplayPasswordPromptReturnsOnCall(0, "", errors.New("current-password-error"))
			cmd.UI = fakeUI
		})

		It("returns the error", func() {
			Expect(executeErr).To(MatchError("current-password-error"))
		})
	})

	When("an error occurs while reading the new password", func() {
		BeforeEach(func() {
			fakeUI.DisplayPasswordPromptReturnsOnCall(1, "", errors.New("new-password-error"))
			cmd.UI = fakeUI
		})

		It("returns the error", func() {
			Expect(executeErr).To(MatchError("new-password-error"))
		})
	})

	When("an error occurs while reading the verification password", func() {
		BeforeEach(func() {
			fakeUI := new(commandfakes.FakeUI)
			fakeUI.DisplayPasswordPromptReturnsOnCall(2, "", errors.New("verify-password-error"))
			cmd.UI = fakeUI
		})

		It("returns the error", func() {
			Expect(executeErr).To(MatchError("verify-password-error"))
		})
	})

	When("the verify password does not match the new password", func() {
		BeforeEach(func() {
			fakeUI.DisplayPasswordPromptReturnsOnCall(2, "WRONG", nil)
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(translatableerror.PasswordVerificationFailedError{}))
			Expect(fakeActor.UpdateUserPasswordCallCount()).To(Equal(0))
		})
	})

	When("an error occurs while making the request to UAA", func() {
		BeforeEach(func() {
			fakeActor.UpdateUserPasswordReturns(errors.New("update-pw-error"))
		})

		It("returns the error", func() {
			Expect(executeErr).To(MatchError("update-pw-error"))
		})
	})
})
