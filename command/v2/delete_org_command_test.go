package v2_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v2"
	"code.cloudfoundry.org/cli/command/v2/v2fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("delete-org Command", func() {
	var (
		cmd             DeleteOrgCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v2fakes.FakeDeleteOrganizationActor
		input           *Buffer
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v2fakes.FakeDeleteOrganizationActor)

		cmd = DeleteOrgCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

		cmd.RequiredArgs.Organization = "some-org"
		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	Context("when a cloud controller API endpoint is set", func() {
		BeforeEach(func() {
			fakeConfig.TargetReturns("some-url")
		})

		Context("when checking target fails", func() {
			BeforeEach(func() {
				fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: binaryName})
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: binaryName}))

				Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
				checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
				Expect(checkTargetedOrg).To(BeFalse())
				Expect(checkTargetedSpace).To(BeFalse())
			})
		})

		Context("when the user is logged in", func() {
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

			Context("when getting the current user does not return an error", func() {
				BeforeEach(func() {
					fakeConfig.CurrentUserReturns(
						configv3.User{Name: "some-user"},
						nil)
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

							Expect(testUI.Out).ToNot(Say("Really delete the org some-org, including its spaces, apps, service instances, routes, private domains and space-scoped service brokers\\? \\[yN\\]:"))
							Expect(testUI.Out).To(Say("Deleting org some-org as some-user..."))

							Expect(fakeActor.DeleteOrganizationCallCount()).To(Equal(1))
							orgName := fakeActor.DeleteOrganizationArgsForCall(0)
							Expect(orgName).To(Equal("some-org"))

							Expect(testUI.Err).To(Say("warning-1"))
							Expect(testUI.Err).To(Say("warning-2"))
							Expect(testUI.Out).To(Say("OK"))
						})
					})

					Context("when an error is encountered deleting the org", func() {
						Context("when the organization does not exist", func() {
							BeforeEach(func() {
								fakeActor.DeleteOrganizationReturns(
									v2action.Warnings{"warning-1", "warning-2"},
									actionerror.OrganizationNotFoundError{
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

							Expect(testUI.Out).To(Say("Really delete the org some-org, including its spaces, apps, service instances, routes, private domains and space-scoped service brokers\\? \\[yN\\]:"))
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

							Expect(testUI.Out).To(Say("Really delete the org some-org, including its spaces, apps, service instances, routes, private domains and space-scoped service brokers\\? \\[yN\\]:"))
							Expect(testUI.Out).To(Say("invalid input \\(not y, n, yes, or no\\)"))
							Expect(testUI.Out).To(Say("Really delete the org some-org, including its spaces, apps, service instances, routes, private domains and space-scoped service brokers\\? \\[yN\\]:"))

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

				Context("when the user deletes the currently targeted org", func() {
					BeforeEach(func() {
						cmd.Force = true
						fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "some-org"})
					})

					It("clears the targeted org and space from the config", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(fakeActor.ClearOrganizationAndSpaceCallCount()).To(Equal(1))
					})
				})

				Context("when the user deletes an org that's not the currently targeted org", func() {
					BeforeEach(func() {
						cmd.Force = true
					})

					It("does not clear the targeted org and space from the config", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(fakeActor.ClearOrganizationAndSpaceCallCount()).To(Equal(0))
					})
				})
			})
		})
	})
})
