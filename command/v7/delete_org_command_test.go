package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
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
		fakeActor       *v7fakes.FakeActor
		input           *Buffer
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		cmd = DeleteOrgCommand{
			BaseCommand: command.BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		cmd.RequiredArgs.Organization = "some-org"
		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("a cloud controller API endpoint is set", func() {
		BeforeEach(func() {
			fakeConfig.TargetReturns("some-url")
		})

		When("checking target fails", func() {
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

		When("the user is logged in", func() {
			When("getting the current user returns an error", func() {
				var returnedErr error

				BeforeEach(func() {
					returnedErr = errors.New("some error")
					fakeConfig.CurrentUserReturns(configv3.User{}, returnedErr)
				})

				It("returns the error", func() {
					Expect(executeErr).To(MatchError(returnedErr))
				})
			})

			When("getting the current user does not return an error", func() {
				BeforeEach(func() {
					fakeConfig.CurrentUserReturns(
						configv3.User{Name: "some-user"},
						nil)
				})

				When("the '-f' flag is provided", func() {
					BeforeEach(func() {
						cmd.Force = true
					})

					When("no errors are encountered", func() {
						BeforeEach(func() {
							fakeActor.DeleteOrganizationReturns(v7action.Warnings{"warning-1", "warning-2"}, nil)
						})

						It("does not prompt for user confirmation, displays warnings, and deletes the org", func() {
							Expect(executeErr).ToNot(HaveOccurred())

							Expect(testUI.Out).ToNot(Say(`Really delete the org some-org, including its spaces, apps, service instances, routes, private domains and space-scoped service brokers\? \[yN\]:`))
							Expect(testUI.Out).To(Say("Deleting org some-org as some-user..."))

							Expect(fakeActor.DeleteOrganizationCallCount()).To(Equal(1))
							orgName := fakeActor.DeleteOrganizationArgsForCall(0)
							Expect(orgName).To(Equal("some-org"))

							Expect(testUI.Err).To(Say("warning-1"))
							Expect(testUI.Err).To(Say("warning-2"))
							Expect(testUI.Out).To(Say("OK"))
						})
					})

					When("an error is encountered deleting the org", func() {
						When("the organization does not exist", func() {
							BeforeEach(func() {
								fakeActor.DeleteOrganizationReturns(
									v7action.Warnings{"warning-1", "warning-2"},
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

								Expect(testUI.Err).To(Say(`Org 'some-org' does not exist\.`))
								Expect(testUI.Out).To(Say("OK"))
							})
						})

						When("the organization does exist", func() {
							var returnedErr error

							BeforeEach(func() {
								returnedErr = errors.New("some error")
								fakeActor.DeleteOrganizationReturns(v7action.Warnings{"warning-1", "warning-2"}, returnedErr)
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
				When("the '-f' flag is not provided", func() {
					When("the user chooses the default", func() {
						BeforeEach(func() {
							_, err := input.Write([]byte("\n"))
							Expect(err).ToNot(HaveOccurred())
						})

						It("does not delete the org", func() {
							Expect(executeErr).ToNot(HaveOccurred())

							Expect(testUI.Out).To(Say(`Organization 'some-org' has not been deleted\.`))

							Expect(fakeActor.DeleteOrganizationCallCount()).To(Equal(0))
						})
					})

					When("the user inputs no", func() {
						BeforeEach(func() {
							_, err := input.Write([]byte("n\n"))
							Expect(err).ToNot(HaveOccurred())
						})

						It("does not delete the org", func() {
							Expect(executeErr).ToNot(HaveOccurred())

							Expect(testUI.Out).To(Say(`Organization 'some-org' has not been deleted\.`))

							Expect(fakeActor.DeleteOrganizationCallCount()).To(Equal(0))
						})
					})

					When("the user inputs yes", func() {
						BeforeEach(func() {
							_, err := input.Write([]byte("y\n"))
							Expect(err).ToNot(HaveOccurred())
						})

						It("deletes the org", func() {
							Expect(executeErr).ToNot(HaveOccurred())

							Expect(testUI.Out).To(Say(`Really delete the org some-org, including its spaces, apps, service instances, routes, private domains and space-scoped service brokers\? \[yN\]:`))
							Expect(testUI.Out).To(Say("Deleting org some-org as some-user..."))

							Expect(fakeActor.DeleteOrganizationCallCount()).To(Equal(1))
							orgName := fakeActor.DeleteOrganizationArgsForCall(0)
							Expect(orgName).To(Equal("some-org"))

							Expect(testUI.Out).To(Say("OK"))
						})
					})

					When("the user input is invalid", func() {
						BeforeEach(func() {
							_, err := input.Write([]byte("e\n\n"))
							Expect(err).ToNot(HaveOccurred())
						})

						It("asks the user again", func() {
							Expect(executeErr).NotTo(HaveOccurred())

							Expect(testUI.Out).To(Say(`Really delete the org some-org, including its spaces, apps, service instances, routes, private domains and space-scoped service brokers\? \[yN\]:`))
							Expect(testUI.Out).To(Say(`invalid input \(not y, n, yes, or no\)`))
							Expect(testUI.Out).To(Say(`Really delete the org some-org, including its spaces, apps, service instances, routes, private domains and space-scoped service brokers\? \[yN\]:`))

							Expect(fakeActor.DeleteOrganizationCallCount()).To(Equal(0))
						})
					})

					When("displaying the prompt returns an error", func() {
						// if nothing is written to input, display bool prompt returns EOF
						It("returns the error", func() {
							Expect(executeErr).To(MatchError("EOF"))
						})
					})
				})

				When("the user deletes the currently targeted org", func() {
					BeforeEach(func() {
						cmd.Force = true
						fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "some-org"})
					})

					It("clears the targeted org and space from the config", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(fakeConfig.UnsetOrganizationAndSpaceInformationCallCount()).To(Equal(1))
					})
				})

				When("the user deletes an org that's not the currently targeted org", func() {
					BeforeEach(func() {
						cmd.Force = true
					})

					It("does not clear the targeted org and space from the config", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(fakeConfig.UnsetOrganizationAndSpaceInformationCallCount()).To(Equal(0))
					})
				})
			})
		})
	})
})
