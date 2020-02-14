package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/translatableerror"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("create-user Command", func() {
	var (
		cmd             CreateUserCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		binaryName      string
		executeErr      error
		input           *Buffer
	)

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		cmd = CreateUserCommand{
			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		cmd.Args.Username = "some-user"
		password := "some-password"
		cmd.Args.Password = &password

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("checking target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: "faceman"}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeFalse())
			Expect(checkTargetedSpace).To(BeFalse())
		})
	})

	When("the user is logged in", func() {
		When("password is not provided", func() {
			BeforeEach(func() {
				cmd.Args.Password = nil
			})

			When("origin is empty string", func() {
				BeforeEach(func() {
					cmd.Origin = ""
				})

				It("returns the RequiredArgumentError", func() {
					Expect(executeErr).To(MatchError(translatableerror.RequiredArgumentError{ArgumentName: "PASSWORD"}))
				})
			})

			When("origin is UAA", func() {
				BeforeEach(func() {
					cmd.Origin = "UAA"
				})

				It("returns the RequiredArgumentError", func() {
					Expect(executeErr).To(MatchError(translatableerror.RequiredArgumentError{ArgumentName: "PASSWORD"}))
				})
			})

			When("origin is not UAA or the empty string", func() {
				BeforeEach(func() {
					fakeActor.CreateUserReturns(
						v7action.User{GUID: "new-user-cc-guid"},
						v7action.Warnings{"warning"},
						nil)
					cmd.Origin = "some-origin"
					fakeActor.GetUserReturns(
						v7action.User{},
						actionerror.UserNotFoundError{})
				})

				It("creates the user and displays all warnings", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(fakeActor.GetUserCallCount()).To(Equal(1))

					username, origin := fakeActor.GetUserArgsForCall(0)
					Expect(username).To(Equal("some-user"))
					Expect(origin).To(Equal("some-origin"))
					Expect(fakeActor.CreateUserCallCount()).To(Equal(1))
					username, password, origin := fakeActor.CreateUserArgsForCall(0)
					Expect(username).To(Equal("some-user"))
					Expect(password).To(Equal(""))
					Expect(origin).To(Equal("some-origin"))

					Expect(testUI.Out).To(Say("Creating user some-user..."))
					Expect(testUI.Out).To(Say("OK"))
					Expect(testUI.Out).To(Say("TIP: Assign roles with 'faceman set-org-role' and 'faceman set-space-role'."))
					Expect(testUI.Err).To(Say("warning"))
				})
			})

			When("password-prompt flag is set", func() {
				BeforeEach(func() {
					cmd.PasswordPrompt = true
					_, err := input.Write([]byte("some-password\n"))
					Expect(err).ToNot(HaveOccurred())

				})

				When("the user already exists in UAA", func() {
					BeforeEach(func() {
						fakeActor.GetUserReturns(
							v7action.User{GUID: "user-guid"},
							nil)
					})

					It("does not prompt for a password or attempt to create a user", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(testUI.Out).To(Not(Say("Password:")))

						Expect(testUI.Out).To(Say("Creating user some-user..."))
						Expect(fakeActor.CreateUserCallCount()).To(Equal(0))
						Expect(testUI.Err).To(Say("User 'some-user' already exists."))
						Expect(testUI.Out).To(Say("OK"))
					})
				})

				When("the user does not yet exist in UAA", func() {
					BeforeEach(func() {
						fakeActor.CreateUserReturns(
							v7action.User{GUID: "new-user-cc-guid"},
							v7action.Warnings{"warning"},
							nil)
						fakeActor.GetUserReturns(
							v7action.User{},
							actionerror.UserNotFoundError{})
					})

					It("prompts for a password", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(testUI.Out).To(Say("Password:"))

						Expect(testUI.Out).To(Say("Creating user some-user..."))

						Expect(fakeActor.CreateUserCallCount()).To(Equal(1))
						_, password, _ := fakeActor.CreateUserArgsForCall(0)
						Expect(password).To(Equal("some-password"))

						Expect(testUI.Out).To(Say("OK"))
						Expect(testUI.Out).To(Say("TIP: Assign roles with 'faceman set-org-role' and 'faceman set-space-role'."))
						Expect(testUI.Err).To(Say("warning"))
					})
				})
			})
		})

		When("password is provided", func() {
			BeforeEach(func() {
				cmd.Args.Username = "some-user"
				password := "password"
				cmd.Args.Password = &password
				cmd.Origin = ""
			})

			When("origin is empty string", func() {
				It("defaults origin to 'uaa'", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(fakeActor.GetUserCallCount()).To(Equal(1))

					username, origin := fakeActor.GetUserArgsForCall(0)
					Expect(username).To(Equal("some-user"))
					Expect(origin).To(Equal(constant.DefaultOriginUaa))
				})
			})
		})

		When("an error occurs", func() {
			When("the error is not translatable", func() {
				var returnedErr error

				BeforeEach(func() {
					returnedErr = errors.New("non-translatable error")
					fakeActor.CreateUserReturns(
						v7action.User{},
						v7action.Warnings{"warning-1", "warning-2"},
						returnedErr)
					fakeActor.GetUserReturns(
						v7action.User{},
						actionerror.UserNotFoundError{})
				})

				It("returns the same error and all warnings", func() {
					Expect(executeErr).To(MatchError(returnedErr))
					Expect(testUI.Err).To(Say("warning-1"))
					Expect(testUI.Err).To(Say("warning-2"))
				})
			})

			When("the error is a uaa.ConflictError", func() {
				var returnedErr error

				BeforeEach(func() {
					returnedErr = uaa.ConflictError{}
					fakeActor.CreateUserReturns(
						v7action.User{},
						v7action.Warnings{"warning-1", "warning-2"},
						returnedErr)
					fakeActor.GetUserReturns(
						v7action.User{},
						actionerror.UserNotFoundError{})
				})

				It("displays the error and all warnings", func() {
					Expect(executeErr).To(BeNil())
					Expect(testUI.Err).To(Say("warning-1"))
					Expect(testUI.Err).To(Say("warning-2"))
					Expect(testUI.Err).To(Say("User 'some-user' already exists."))
				})
			})
		})
	})
})
