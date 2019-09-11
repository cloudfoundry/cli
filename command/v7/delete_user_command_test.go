package v7_test

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("delete-user Command", func() {
	var (
		cmd             DeleteUserCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeDeleteUserActor
		binaryName      string
		executeErr      error
		input           *Buffer
		currentUser     string
	)

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeDeleteUserActor)
		currentUser, _ = fakeConfig.CurrentUserName()

		cmd = DeleteUserCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

		cmd.RequiredArgs.Username = "some-user"

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
		When("no errors occur", func() {
			BeforeEach(func() {
				cmd.Origin = "some-origin"
				fakeActor.GetUserReturns(v7action.User{
					GUID: "some-user-guid", Origin: "some-origin",
				}, nil)
				fakeActor.DeleteUserReturns(v7action.Warnings{"warning: user is about to be deleted"}, nil)
			})

			When("the -f flag is provided", func() {
				BeforeEach(func() {
					cmd.Force = true
				})

				It("deletes the user", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(fakeActor.GetUserCallCount()).To(Equal(1))

					username, origin := fakeActor.GetUserArgsForCall(0)
					Expect(username).To(Equal("some-user"))
					Expect(origin).To(Equal("some-origin"))

					Expect(fakeActor.DeleteUserCallCount()).To(Equal(1))
					userGuid := fakeActor.DeleteUserArgsForCall(0)
					Expect(userGuid).To(Equal("some-user-guid"))

					Expect(testUI.Out).To(Say(`Deleting user some-user as %s\.\.\.`, currentUser))
					Expect(testUI.Out).To(Say("OK"))
					Expect(testUI.Err).To(Say("warning: user is about to be deleted"))
				})
			})

			When("the -f flag is NOT provided", func() {
				BeforeEach(func() {
					cmd.Force = false
				})

				When("the user inputs yes", func() {
					BeforeEach(func() {
						_, err := input.Write([]byte("y\n"))
						Expect(err).ToNot(HaveOccurred())
					})

					It("deletes the user", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say(`Really delete the user some-user\? \[yN\]`))

						username, origin := fakeActor.GetUserArgsForCall(0)
						Expect(username).To(Equal("some-user"))
						Expect(origin).To(Equal("some-origin"))

						Expect(fakeActor.DeleteUserCallCount()).To(Equal(1))
						userGuid := fakeActor.DeleteUserArgsForCall(0)
						Expect(userGuid).To(Equal("some-user-guid"))

						Expect(testUI.Out).To(Say("Deleting user some-user as %s...", currentUser))
						Expect(testUI.Out).To(Say("OK"))
						Expect(testUI.Err).To(Say("warning: user is about to be deleted"))
					})
				})

				When("the user inputs no", func() {
					BeforeEach(func() {
						_, err := input.Write([]byte("n\n"))
						Expect(err).ToNot(HaveOccurred())
					})

					It("deletes the user", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say(`Really delete the user some-user\? \[yN\]`))
						Expect(fakeActor.DeleteUserCallCount()).To(Equal(0))

						Expect(testUI.Out).NotTo(Say(`Deleting user some-user as %s\.\.\.`, currentUser))
						Expect(testUI.Err).NotTo(Say("warning: user is about to be deleted"))
						Expect(testUI.Out).To(Say(`User 'some-user' has not been deleted.`))
						Expect(testUI.Out).To(Say("OK"))
					})
				})

				When("the user chooses the default", func() {
					BeforeEach(func() {
						_, err := input.Write([]byte("\n"))
						Expect(err).ToNot(HaveOccurred())
					})

					It("deletes the user", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say(`Really delete the user some-user\? \[yN\]`))
						Expect(fakeActor.DeleteUserCallCount()).To(Equal(0))

						Expect(testUI.Out).NotTo(Say(`Deleting user some-user as %s\.\.\.`, currentUser))
						Expect(testUI.Err).NotTo(Say("warning: user is about to be deleted"))
						Expect(testUI.Out).To(Say(`User 'some-user' has not been deleted.`))
						Expect(testUI.Out).To(Say("OK"))
					})
				})
			})
		})

		When("an error occurs", func() {
			BeforeEach(func() {
				cmd.Force = true
			})

			When("GetUser action errors", func() {
				When("no user is found", func() {
					var returnedErr error

					BeforeEach(func() {
						returnedErr = actionerror.UAAUserNotFoundError{Username: "some-user"}
						fakeActor.GetUserReturns(
							v7action.User{},
							returnedErr)
					})

					It("returns the same error", func() {
						Expect(executeErr).NotTo(HaveOccurred())
						Expect(testUI.Out).To(Say(`OK`))
						Expect(testUI.Out).To(Say(`User 'some-user' does not exist.`))
					})
				})
			})

			When("DeleteUser action errors", func() {
				var returnedErr error

				BeforeEach(func() {
					returnedErr = uaa.ConflictError{}
					fakeActor.GetUserReturns(
						v7action.User{GUID: "some-guid", Origin: "uaa"},
						nil)
					fakeActor.DeleteUserReturns(nil, returnedErr)
				})

				It("returns the same error", func() {
					Expect(executeErr).To(MatchError(returnedErr))
				})
			})

			When("when everything succeeds", func() {
				var returnedErr error

				BeforeEach(func() {
					fakeActor.GetUserReturns(
						v7action.User{GUID: "some-guid", Origin: "uaa"},
						nil)
					warnings := []string{"warning-1", "warning-2"}
					fakeActor.DeleteUserReturns(warnings, returnedErr)
				})

				It("displays all warnings", func() {
					Expect(executeErr).To(BeNil())
					Expect(testUI.Err).To(Say("warning-1"))
					Expect(testUI.Err).To(Say("warning-2"))
				})
			})
		})
	})
})
