package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/ui"
	"github.com/SermoDigital/jose/crypto"
	"github.com/SermoDigital/jose/jws"
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
		fakeActor       *v7fakes.FakeActor
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		cmd = OauthTokenCommand{
			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("checking the target fails", func() {
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

	When("logged in as a client", func() {
		BeforeEach(func() {
			fakeConfig.UAAGrantTypeReturns("client_credentials")
		})

		When("the existing access token is invalid", func() {
			BeforeEach(func() {
				token := jws.NewJWT(jws.Claims{}, crypto.SigningMethodHS256)
				fakeConfig.AccessTokenReturns("invalid-existing-access-token")
				fakeActor.ParseAccessTokenReturns(token, errors.New("Access token is invalid"))
			})

			It("errors", func() {
				Expect(executeErr).To(MatchError(errors.New("Access token is invalid.")))

				Expect(testUI.Out).ToNot(Say("new-access-token"))

				Expect(fakeActor.RefreshAccessTokenCallCount()).To(Equal(0))
				Expect(fakeActor.ParseAccessTokenCallCount()).To(Equal(1))
				Expect(fakeActor.ParseAccessTokenArgsForCall(0)).To(Equal("invalid-existing-access-token"))
			})
		})

		When("the existing access token does not have an expiry time", func() {
			BeforeEach(func() {
				token := jws.NewJWT(jws.Claims{}, crypto.SigningMethodHS256)
				fakeConfig.AccessTokenReturns("existing-access-token")
				fakeActor.ParseAccessTokenReturns(token, nil)
			})

			It("errors", func() {
				Expect(executeErr).To(MatchError(errors.New("Access token is missing expiration claim.")))

				Expect(testUI.Out).ToNot(Say("new-access-token"))

				Expect(fakeActor.RefreshAccessTokenCallCount()).To(Equal(0))
				Expect(fakeActor.ParseAccessTokenCallCount()).To(Equal(1))
				Expect(fakeActor.ParseAccessTokenArgsForCall(0)).To(Equal("existing-access-token"))
			})
		})

	})

	When("logged in as a user", func() {
		BeforeEach(func() {
			fakeConfig.RefreshTokenReturns("existing-refresh-token")
		})

		When("an error is encountered refreshing the access token", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("refresh access token error")
				fakeActor.RefreshAccessTokenReturns("", expectedErr)
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError(expectedErr))

				Expect(testUI.Out).ToNot(Say("new-access-token"))

				Expect(fakeActor.RefreshAccessTokenCallCount()).To(Equal(1))
			})
		})

		When("no errors are encountered refreshing the access token", func() {
			BeforeEach(func() {
				fakeActor.RefreshAccessTokenReturns("new-access-token", nil)
			})

			It("refreshes the access and refresh tokens and displays the access token", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Out).To(Say("new-access-token"))

				Expect(fakeActor.RefreshAccessTokenCallCount()).To(Equal(1))
			})
		})
	})
})
