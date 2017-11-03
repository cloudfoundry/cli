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

var _ = Describe("delete-space Command", func() {
	var (
		cmd             DeleteSpaceCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v2fakes.FakeDeleteSpaceActor
		input           *Buffer
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v2fakes.FakeDeleteSpaceActor)

		cmd = DeleteSpaceCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

		cmd.RequiredArgs.Space = "some-space"

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		fakeConfig.CurrentUserReturns(configv3.User{Name: "some-user"}, nil)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	Context("when a cloud controller API endpoint is set", func() {
		BeforeEach(func() {
			fakeConfig.TargetReturns("some-url")
		})

		Context("when checking target fails", func() {
			Context("when an org is provided", func() {
				BeforeEach(func() {
					cmd.Org = "some-org"
					fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: binaryName})
				})

				It("returns the NotLoggedInError", func() {
					Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: binaryName}))

					checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
					Expect(checkTargetedOrg).To(BeFalse())
					Expect(checkTargetedSpace).To(BeFalse())
				})
			})

			Context("when an org is NOT provided", func() {
				BeforeEach(func() {
					fakeSharedActor.CheckTargetReturns(actionerror.NoOrganizationTargetedError{})
				})

				It("returns the NoOrganizationTargetedError", func() {
					Expect(executeErr).To(MatchError(actionerror.NoOrganizationTargetedError{}))

					checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
					Expect(checkTargetedOrg).To(BeTrue())
					Expect(checkTargetedSpace).To(BeFalse())
				})
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

			Context("when the -o flag is provided", func() {
				BeforeEach(func() {
					cmd.Org = "some-org"
				})

				Context("when the -f flag is provided", func() {
					BeforeEach(func() {
						cmd.Force = true
					})

					Context("when the deleting the space errors", func() {
						BeforeEach(func() {
							fakeActor.DeleteSpaceByNameAndOrganizationNameReturns(v2action.Warnings{"warning-1", "warning-2"}, actionerror.SpaceNotFoundError{Name: "some-space"})
						})

						It("returns the translatable error", func() {
							Expect(executeErr).To(MatchError(actionerror.SpaceNotFoundError{Name: "some-space"}))
							Expect(testUI.Out).To(Say("Deleting space some-space in org some-org as some-user\\.\\.\\."))

							Expect(testUI.Err).To(Say("warning-1"))
							Expect(testUI.Err).To(Say("warning-2"))
						})
					})

					Context("when the deleting the space succeeds", func() {
						BeforeEach(func() {
							fakeActor.DeleteSpaceByNameAndOrganizationNameReturns(v2action.Warnings{"warning-1", "warning-2"}, nil)
						})

						Context("when the user was targeted to the space", func() {
							BeforeEach(func() {
								fakeConfig.TargetedSpaceReturns(configv3.Space{Name: "some-space"})
								fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "some-org"})
							})

							It("untargets the space, displays all warnings and does not error", func() {
								Expect(executeErr).ToNot(HaveOccurred())

								Expect(testUI.Out).To(Say("Deleting space some-space in org some-org as some-user\\.\\.\\."))
								Expect(testUI.Out).To(Say("OK"))
								Expect(testUI.Out).To(Say("TIP: No space targeted, use 'faceman target -s' to target a space."))

								Expect(testUI.Err).To(Say("warning-1"))
								Expect(testUI.Err).To(Say("warning-2"))

								Expect(fakeConfig.UnsetSpaceInformationCallCount()).To(Equal(1))

								spaceArg, orgArg := fakeActor.DeleteSpaceByNameAndOrganizationNameArgsForCall(0)
								Expect(spaceArg).To(Equal("some-space"))
								Expect(orgArg).To(Equal("some-org"))
							})
						})

						Context("when the user was NOT targeted to the space", func() {
							BeforeEach(func() {
								fakeConfig.TargetedSpaceReturns(configv3.Space{Name: "some-space"})
								fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "some-other-org"})
							})

							It("displays all warnings and does not error", func() {
								Expect(executeErr).ToNot(HaveOccurred())

								Expect(testUI.Out).To(Say("Deleting space some-space in org some-org as some-user\\.\\.\\."))
								Expect(testUI.Out).To(Say("OK"))
								Expect(testUI.Out).ToNot(Say("TIP: No space targeted, use 'faceman target -s' to target a space."))

								Expect(testUI.Err).To(Say("warning-1"))
								Expect(testUI.Err).To(Say("warning-2"))

								Expect(fakeConfig.UnsetSpaceInformationCallCount()).To(Equal(0))
							})
						})
					})
				})

				Context("when the -f flag is NOT provided", func() {
					BeforeEach(func() {
						cmd.Force = false
					})

					Context("when the user inputs yes", func() {
						BeforeEach(func() {
							_, err := input.Write([]byte("y\n"))
							Expect(err).ToNot(HaveOccurred())

							fakeActor.DeleteSpaceByNameAndOrganizationNameReturns(v2action.Warnings{"warning-1", "warning-2"}, nil)
						})

						It("deletes the space", func() {
							Expect(executeErr).ToNot(HaveOccurred())

							Expect(testUI.Out).To(Say("Really delete the space some-space\\? \\[yN\\]"))
							Expect(testUI.Out).To(Say("Deleting space some-space in org some-org as some-user\\.\\.\\."))
							Expect(testUI.Out).To(Say("OK"))
							Expect(testUI.Out).ToNot(Say("TIP: No space targeted, use 'faceman target -s' to target a space."))

							Expect(testUI.Err).To(Say("warning-1"))
							Expect(testUI.Err).To(Say("warning-2"))
						})
					})

					Context("when the user inputs no", func() {
						BeforeEach(func() {
							_, err := input.Write([]byte("n\n"))
							Expect(err).ToNot(HaveOccurred())
						})

						It("cancels the delete", func() {
							Expect(executeErr).ToNot(HaveOccurred())

							Expect(testUI.Out).To(Say("Delete cancelled"))
							Expect(fakeActor.DeleteSpaceByNameAndOrganizationNameCallCount()).To(Equal(0))
						})
					})

					Context("when the user chooses the default", func() {
						BeforeEach(func() {
							_, err := input.Write([]byte("\n"))
							Expect(err).ToNot(HaveOccurred())
						})

						It("cancels the delete", func() {
							Expect(executeErr).ToNot(HaveOccurred())

							Expect(testUI.Out).To(Say("Delete cancelled"))
							Expect(fakeActor.DeleteSpaceByNameAndOrganizationNameCallCount()).To(Equal(0))
						})
					})

					Context("when the user input is invalid", func() {
						BeforeEach(func() {
							_, err := input.Write([]byte("e\n\n"))
							Expect(err).ToNot(HaveOccurred())
						})

						It("asks the user again", func() {
							Expect(executeErr).NotTo(HaveOccurred())

							Expect(testUI.Out).To(Say("Really delete the space some-space\\? \\[yN\\]"))
							Expect(testUI.Out).To(Say("invalid input \\(not y, n, yes, or no\\)"))
							Expect(testUI.Out).To(Say("Really delete the space some-space\\? \\[yN\\]"))

							Expect(fakeActor.DeleteSpaceByNameAndOrganizationNameCallCount()).To(Equal(0))
						})
					})
				})
			})

			Context("when the -o flag is NOT provided", func() {
				BeforeEach(func() {
					cmd.Org = ""
					cmd.Force = true
					fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "some-targeted-org"})
					fakeActor.DeleteSpaceByNameAndOrganizationNameReturns(v2action.Warnings{"warning-1", "warning-2"}, nil)
				})

				It("deletes the space in the targeted org", func() {
					Expect(executeErr).NotTo(HaveOccurred())

					Expect(testUI.Out).To(Say("Deleting space some-space in org some-targeted-org as some-user\\.\\.\\."))
					Expect(testUI.Out).To(Say("OK"))
					Expect(testUI.Out).ToNot(Say("TIP: No space targeted, use 'faceman target -s' to target a space\\."))

					Expect(testUI.Err).To(Say("warning-1"))
					Expect(testUI.Err).To(Say("warning-2"))

					spaceArg, orgArg := fakeActor.DeleteSpaceByNameAndOrganizationNameArgsForCall(0)
					Expect(spaceArg).To(Equal("some-space"))
					Expect(orgArg).To(Equal("some-targeted-org"))
				})

				Context("when deleting a targeted space", func() {
					BeforeEach(func() {
						fakeConfig.TargetedSpaceReturns(configv3.Space{Name: "some-space"})
					})

					It("deletes the space and untargets the org", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say("Deleting space some-space in org some-targeted-org as some-user\\.\\.\\."))
						Expect(testUI.Out).To(Say("OK"))
						Expect(testUI.Out).To(Say("TIP: No space targeted, use 'faceman target -s' to target a space."))

						Expect(testUI.Err).To(Say("warning-1"))
						Expect(testUI.Err).To(Say("warning-2"))

						Expect(fakeConfig.UnsetSpaceInformationCallCount()).To(Equal(1))
					})
				})
			})
		})
	})
})
