package v2_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v2"
	"code.cloudfoundry.org/cli/command/v2/v2fakes"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("oauth-token command", func() {
	var (
		cmd             OauthTokenCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v2fakes.FakeOauthTokenActor
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v2fakes.FakeOauthTokenActor)

		cmd = OauthTokenCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	Context("when checking the target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: binaryName})
		})

		It("returns a wrapped error", func() {
			Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: binaryName}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargettedOrgArg, checkTargettedSpaceArg := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargettedOrgArg).To(BeFalse())
			Expect(checkTargettedSpaceArg).To(BeFalse())
		})
	})

	Context("when the user is logged in", func() {
		BeforeEach(func() {
			fakeConfig.RefreshTokenReturns("existing-refresh-token")
		})

		Context("when an error is encountered refreshing the access token", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("refresh access token error")
				fakeActor.RefreshAccessTokenReturns("", expectedErr)
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError(expectedErr))

				Expect(testUI.Out).ToNot(Say("new-access-token"))

				Expect(fakeActor.RefreshAccessTokenCallCount()).To(Equal(1))
				Expect(fakeActor.RefreshAccessTokenArgsForCall(0)).To(Equal("existing-refresh-token"))
			})
		})

		Context("when no errors are encountered refreshing the access token", func() {
			BeforeEach(func() {
				fakeActor.RefreshAccessTokenReturns("new-access-token", nil)
			})

			It("refreshes the access and refresh tokens and displays the access token", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Out).To(Say("new-access-token"))

				Expect(fakeActor.RefreshAccessTokenCallCount()).To(Equal(1))
				Expect(fakeActor.RefreshAccessTokenArgsForCall(0)).To(Equal("existing-refresh-token"))
			})
		})
	})
})
