package v2_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/v2"
	"code.cloudfoundry.org/cli/command/v2/v2fakes"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("CreateUser Command", func() {
	var (
		cmd        v2.CreateUserCommand
		fakeUI     *ui.UI
		fakeConfig *commandfakes.FakeConfig
		fakeActor  *v2fakes.FakeCreateUserActor
		executeErr error
	)

	BeforeEach(func() {
		out := NewBuffer()
		fakeUI = ui.NewTestUI(nil, out, out)
		fakeConfig = new(commandfakes.FakeConfig)
		fakeActor = new(v2fakes.FakeCreateUserActor)

		cmd = v2.CreateUserCommand{
			UI:     fakeUI,
			Config: fakeConfig,
			Actor:  fakeActor,
		}

		cmd.RequiredArgs.Username = "some-user"
		cmd.RequiredArgs.Password = "some-password"

		fakeConfig.ExperimentalReturns(true)
		fakeConfig.BinaryNameReturns("faceman")
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	It("Displays the experimental warning message", func() {
		Expect(fakeUI.Out).To(Say(command.ExperimentalWarning))
	})

	Context("when the user is not logged in", func() {
		BeforeEach(func() {
			fakeConfig.AccessTokenReturns("")
			fakeConfig.RefreshTokenReturns("")
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(command.NotLoggedInError{
				BinaryName: "faceman",
			}))
		})
	})

	Context("when the user is logged in", func() {
		BeforeEach(func() {
			fakeConfig.AccessTokenReturns("some-access-token")
			fakeConfig.RefreshTokenReturns("some-refresh-token")
		})

		Context("when no errors occur", func() {
			BeforeEach(func() {
				fakeActor.NewUserReturns(
					v2action.User{
						GUID: "new-user-cc-guid",
					},
					v2action.Warnings{
						"user already exists",
						"warning-2",
					},
					nil,
				)
			})

			It("creates the user and displays all warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeActor.NewUserCallCount()).To(Equal(1))
				username, password := fakeActor.NewUserArgsForCall(0)
				Expect(username).To(Equal("some-user"))
				Expect(password).To(Equal("some-password"))

				Expect(fakeUI.Out).To(Say(`
Creating user some-user...
user already exists
warning-2
OK

TIP: Assign roles with 'faceman set-org-role' and 'faceman set-space-role'.`))
			})
		})

		Context("when an error occurs", func() {
			Context("when the error is not translatable", func() {
				var returnedErr error

				BeforeEach(func() {
					returnedErr = errors.New("non-translatable error")
					fakeActor.NewUserReturns(
						v2action.User{},
						v2action.Warnings{
							"warning-1",
							"warning-2",
						},
						returnedErr,
					)
				})

				It("returns the same error and all warnings", func() {
					Expect(executeErr).To(MatchError(returnedErr))
					Expect(fakeUI.Err).To(Say("warning-1"))
					Expect(fakeUI.Err).To(Say("warning-2"))
				})
			})
		})
	})
})
