package v2_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/v2"
	"code.cloudfoundry.org/cli/command/v2/v2fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("delete-org Command", func() {
	var (
		cmd        v2.DeleteOrgCommand
		testUI     *ui.UI
		fakeActor  *v2fakes.FakeDeleteOrganizationActor
		fakeConfig *commandfakes.FakeConfig
		input      *Buffer
		executeErr error
	)

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeActor = new(v2fakes.FakeDeleteOrganizationActor)
		fakeConfig = new(commandfakes.FakeConfig)

		cmd = v2.DeleteOrgCommand{
			UI:     testUI,
			Actor:  fakeActor,
			Config: fakeConfig,
		}

		cmd.RequiredArgs.Organization = "some-org"
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	Context("when the user is not logged in", func() {
		BeforeEach(func() {
			fakeConfig.BinaryNameReturns("faceman")
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(command.NotLoggedInError{
				BinaryName: "faceman",
			}))
		})
	})

	Context("when the user is logged in", func() {
		BeforeEach(func() {
			fakeConfig.TargetReturns("some-url")
			fakeConfig.AccessTokenReturns("some-access-token")
			fakeConfig.RefreshTokenReturns("some-refresh-token")
			fakeConfig.CurrentUserReturns(configv3.User{
				Name: "some-user",
			}, nil)
		})

		Context("when the '-f' flag is provided", func() {
			BeforeEach(func() {
				cmd.Force = true
			})

			Context("when no errors are encountered", func() {
				BeforeEach(func() {
					fakeActor.DeleteOrganizationReturns(v2action.Warnings{"warning-1", "warning-2"}, nil)
				})

				It("does not prompt for user confirmation, displays warnings, and deletes the org", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).ToNot(Say("Really delete the org some-org and everything associated with it\\?>> \\[yN\\]:"))
					Expect(testUI.Out).To(Say("Deleting org some-org as some-user..."))

					Expect(fakeActor.DeleteOrganizationCallCount()).To(Equal(1))
					orgName := fakeActor.DeleteOrganizationArgsForCall(0)
					Expect(orgName).To(Equal("some-org"))

					Expect(testUI.Err).To(Say("warning-1"))
					Expect(testUI.Err).To(Say("warning-2"))
					Expect(testUI.Out).To(Say("OK"))
				})
			})

			Context("when getting the current user returns an error", func() {
				var returnedErr error

				BeforeEach(func() {
					returnedErr = errors.New("some error")
					fakeConfig.CurrentUserReturns(configv3.User{}, returnedErr)
				})

				It("returns the error", func() {
					Expect(executeErr).To(MatchError(returnedErr))
				})
			})

			Context("when the delete actor org returns an error", func() {
				Context("when the organization does not exist", func() {
					BeforeEach(func() {
						fakeActor.DeleteOrganizationReturns(
							v2action.Warnings{"warning-1", "warning-2"},
							v2action.OrganizationNotFoundError{
								Name: "some-org",
							},
						)
					})

					It("returns an OrganizationNotFoundError and displays all warnings", func() {
						Expect(executeErr).NotTo(HaveOccurred())

						Expect(testUI.Out).To(Say("Deleting org some-org as some-user..."))

						Expect(fakeActor.DeleteOrganizationCallCount()).To(Equal(1))
						orgName := fakeActor.DeleteOrganizationArgsForCall(0)
						Expect(orgName).To(Equal("some-org"))

						Expect(testUI.Err).To(Say("warning-1"))
						Expect(testUI.Err).To(Say("warning-2"))

						Expect(testUI.Out).To(Say("Org some-org does not exist."))
						Expect(testUI.Out).To(Say("OK"))
					})
				})

				Context("when the organization does exist", func() {
					var returnedErr error

					BeforeEach(func() {
						returnedErr = errors.New("some error")
						fakeActor.DeleteOrganizationReturns(v2action.Warnings{"warning-1", "warning-2"}, returnedErr)
					})

					It("returns the error, displays all warnings, and does not delete the org", func() {
						Expect(executeErr).To(MatchError(returnedErr))

						Expect(testUI.Out).To(Say("Deleting org some-org as some-user..."))

						Expect(fakeActor.DeleteOrganizationCallCount()).To(Equal(1))
						orgName := fakeActor.DeleteOrganizationArgsForCall(0)
						Expect(orgName).To(Equal("some-org"))

						Expect(testUI.Err).To(Say("warning-1"))
						Expect(testUI.Err).To(Say("warning-2"))
					})
				})
			})
		})

		// Testing the prompt.
		Context("when the '-f' flag is not provided", func() {
			Context("when the user chooses the default", func() {
				BeforeEach(func() {
					input.Write([]byte("\n"))
				})

				It("does not delete the org", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say("Delete cancelled"))

					Expect(fakeActor.DeleteOrganizationCallCount()).To(Equal(0))
				})
			})

			Context("when the user inputs no", func() {
				BeforeEach(func() {
					input.Write([]byte("n\n"))
				})

				It("does not delete the org", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say("Delete cancelled"))

					Expect(fakeActor.DeleteOrganizationCallCount()).To(Equal(0))
				})
			})

			Context("when the user inputs yes", func() {
				BeforeEach(func() {
					input.Write([]byte("y\n"))
				})

				It("deletes the org", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say("Really delete the org some-org and everything associated with it\\?>> \\[yN\\]:"))
					Expect(testUI.Out).To(Say("Deleting org some-org as some-user..."))

					Expect(fakeActor.DeleteOrganizationCallCount()).To(Equal(1))
					orgName := fakeActor.DeleteOrganizationArgsForCall(0)
					Expect(orgName).To(Equal("some-org"))

					Expect(testUI.Out).To(Say("OK"))
				})
			})

			Context("when the user input is invalid", func() {
				BeforeEach(func() {
					input.Write([]byte("e\n\n"))
				})

				It("asks the user again", func() {
					Expect(executeErr).NotTo(HaveOccurred())

					Expect(testUI.Out).To(Say("Really delete the org some-org and everything associated with it\\?>> \\[yN\\]:"))
					Expect(testUI.Out).To(Say("invalid input \\(not y, n, yes, or no\\)"))
					Expect(testUI.Out).To(Say("Really delete the org some-org and everything associated with it\\?>> \\[yN\\]:"))

					Expect(fakeActor.DeleteOrganizationCallCount()).To(Equal(0))
				})
			})

			Context("when displaying the prompt returns an error", func() {
				// if nothing is written to input, display bool prompt returns EOF
				It("returns the error", func() {
					Expect(executeErr).To(MatchError("EOF"))
				})
			})
		})
	})
})
