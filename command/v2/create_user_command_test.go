package v2_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/api/uaa"
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
		testUI     *ui.UI
		fakeConfig *commandfakes.FakeConfig
		fakeActor  *v2fakes.FakeCreateUserActor
		executeErr error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeActor = new(v2fakes.FakeCreateUserActor)

		cmd = v2.CreateUserCommand{
			UI:     testUI,
			Config: fakeConfig,
			Actor:  fakeActor,
		}

		cmd.Args.Username = "some-user"
		password := "some-password"
		cmd.Args.Password = &password

		fakeConfig.BinaryNameReturns("faceman")
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
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

		Context("when password is not provided", func() {
			BeforeEach(func() {
				cmd.Args.Password = nil
			})

			Context("when origin is empty string", func() {
				BeforeEach(func() {
					cmd.Origin = ""
				})
				It("returns the RequiredArgumentError", func() {
					Expect(executeErr).To(MatchError(command.RequiredArgumentError{
						ArgumentName: "PASSWORD",
					}))
				})
			})

			Context("when origin is UAA", func() {
				BeforeEach(func() {
					cmd.Origin = "UAA"
				})
				It("returns the RequiredArgumentError", func() {
					Expect(executeErr).To(MatchError(command.RequiredArgumentError{
						ArgumentName: "PASSWORD",
					}))
				})
			})

			Context("when origin is not UAA or the empty string", func() {
				BeforeEach(func() {
					fakeActor.NewUserReturns(
						v2action.User{
							GUID: "new-user-cc-guid",
						},
						v2action.Warnings{
							"warning",
						},
						nil,
					)
					cmd.Origin = "some-origin"
				})
				It("creates the user and displays all warnings", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(fakeActor.NewUserCallCount()).To(Equal(1))
					username, password, origin := fakeActor.NewUserArgsForCall(0)
					Expect(username).To(Equal("some-user"))
					Expect(password).To(Equal(""))
					Expect(origin).To(Equal("some-origin"))

					Expect(testUI.Out).To(Say("Creating user some-user..."))
					Expect(testUI.Out).To(Say("OK"))
					Expect(testUI.Out).To(Say("TIP: Assign roles with 'faceman set-org-role' and 'faceman set-space-role'."))
					Expect(testUI.Err).To(Say("warning"))
				})
			})
		})

		Context("when no errors occur", func() {
			BeforeEach(func() {
				fakeActor.NewUserReturns(
					v2action.User{
						GUID: "new-user-cc-guid",
					},
					v2action.Warnings{
						"warning",
					},
					nil,
				)
				cmd.Origin = "some-origin"
			})

			It("creates the user and displays all warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeActor.NewUserCallCount()).To(Equal(1))
				username, password, origin := fakeActor.NewUserArgsForCall(0)
				Expect(username).To(Equal("some-user"))
				Expect(password).To(Equal("some-password"))
				Expect(origin).To(Equal("some-origin"))

				Expect(testUI.Out).To(Say("Creating user some-user..."))
				Expect(testUI.Out).To(Say("OK"))
				Expect(testUI.Out).To(Say("TIP: Assign roles with 'faceman set-org-role' and 'faceman set-space-role'."))
				Expect(testUI.Err).To(Say("warning"))
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
					Expect(testUI.Err).To(Say("warning-1"))
					Expect(testUI.Err).To(Say("warning-2"))
				})
			})

			Context("when the error is a uaa.ConflictError", func() {
				var returnedErr error

				BeforeEach(func() {
					returnedErr = uaa.ConflictError{}
					fakeActor.NewUserReturns(
						v2action.User{},
						v2action.Warnings{
							"warning-1",
							"warning-2",
						},
						returnedErr,
					)
				})

				It("displays the error and all warnings", func() {
					Expect(executeErr).To(BeNil())
					Expect(testUI.Err).To(Say("warning-1"))
					Expect(testUI.Err).To(Say("warning-2"))
					Expect(testUI.Err).To(Say("user some-user already exists"))
				})
			})
		})
	})
})
