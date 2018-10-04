package v6_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	. "code.cloudfoundry.org/cli/command/v6"
	"code.cloudfoundry.org/cli/command/v6/v6fakes"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("CreateSpaceCommand", func() {
	var (
		fakeConfig      *commandfakes.FakeConfig
		fakeActor       *v6fakes.FakeCreateSpaceActor
		fakeSharedActor *commandfakes.FakeSharedActor
		testUI          *ui.UI
		spaceName       string
		cmd             CreateSpaceCommand

		executeErr error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeActor = new(v6fakes.FakeCreateSpaceActor)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		spaceName = "some-space"

		cmd = CreateSpaceCommand{
			UI:           testUI,
			Config:       fakeConfig,
			Actor:        fakeActor,
			SharedActor:  fakeSharedActor,
			RequiredArgs: flag.Space{Space: spaceName},
		}
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	It("checks for user being logged in", func() {
		Expect(fakeSharedActor.RequireCurrentUserCallCount()).To(Equal(1))
	})

	When("user is not logged in", func() {
		expectedErr := errors.New("not logged in and/or can't verify login because of error")

		BeforeEach(func() {
			fakeSharedActor.RequireCurrentUserReturns("", expectedErr)
		})

		It("returns the error", func() {
			Expect(executeErr).To(MatchError(expectedErr))
		})
	})

	When("user is logged in", func() {
		var username string

		BeforeEach(func() {
			username = "some-guy"
			fakeSharedActor.RequireCurrentUserReturns(username, nil)
		})

		When("user specifies an org using the -o flag", func() {
			var specifiedOrgName string

			BeforeEach(func() {
				specifiedOrgName = "specified-org"
				cmd.Organization = specifiedOrgName
			})

			It("does not require a targeted org", func() {
				Expect(fakeSharedActor.RequireTargetedOrgCallCount()).To(Equal(0))
			})

			It("creates a space in the specified org with no errors", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(testUI.Out).To(Say(`Creating space %s in org %s as %s\.\.\.`, spaceName, specifiedOrgName, username))

				Expect(testUI.Out).To(Say("OK\n\n"))

				Expect(fakeActor.CreateSpaceCallCount()).To(Equal(1))
				inputSpace, inputOrg, _ := fakeActor.CreateSpaceArgsForCall(0)
				Expect(inputSpace).To(Equal(spaceName))
				Expect(inputOrg).To(Equal(specifiedOrgName))
			})

			When("an org is targeted", func() {
				BeforeEach(func() {
					fakeSharedActor.RequireTargetedOrgReturns("do-not-use-this-org", nil)
				})

				It("uses the specified org, not the targeted org", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(testUI.Out).To(Say(`Creating space %s in org %s as %s\.\.\.`, spaceName, specifiedOrgName, username))
					_, inputOrg, _ := fakeActor.CreateSpaceArgsForCall(0)
					Expect(inputOrg).To(Equal(specifiedOrgName))
				})
			})
		})

		When("no org is specified using the -o flag", func() {
			It("requires a targeted org", func() {
				Expect(fakeSharedActor.RequireTargetedOrgCallCount()).To(Equal(1))
			})

			When("no org is targeted", func() {
				BeforeEach(func() {
					fakeSharedActor.RequireTargetedOrgReturns("", errors.New("check target error"))
				})

				It("returns an error", func() {
					Expect(executeErr).To(MatchError("check target error"))
				})
			})

			When("an org is targeted", func() {
				var orgName string

				BeforeEach(func() {
					fakeSharedActor.RequireTargetedOrgReturns(orgName, nil)
				})

				It("attempts to create a space with the specified name in the targeted org", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(fakeActor.CreateSpaceCallCount()).To(Equal(1))
					inputSpace, inputOrg, _ := fakeActor.CreateSpaceArgsForCall(0)
					Expect(inputSpace).To(Equal(spaceName))
					Expect(inputOrg).To(Equal(orgName))
				})

				When("creating the space succeeds", func() {
					BeforeEach(func() {
						fakeActor.CreateSpaceReturns(
							v2action.Space{GUID: "some-space-guid", OrganizationGUID: "some-org-guid"},
							v2action.Warnings{"warn-1", "warn-2"},
							nil,
						)
					})

					It("tells you that it plans to give the user space roles", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(testUI.Out).To(Say(`Creating space %s in org %s as %s\.\.\.`, spaceName, orgName, username))

						Expect(testUI.Err).To(Say("warn-1\nwarn-2\n"))
						Expect(testUI.Out).To(Say("OK\n"))

						Expect(testUI.Out).To(Say(`Assigning role SpaceManager to user %s in org %s / space %s as %s\.\.\.`, username, orgName, spaceName, username))
						Expect(testUI.Out).To(Say("OK\n"))
						Expect(testUI.Out).To(Say(`Assigning role SpaceDeveloper to user %s in org %s / space %s as %s\.\.\.`, username, orgName, spaceName, username))
						Expect(testUI.Out).To(Say("OK\n\n"))
						Eventually(testUI.Out).Should(Say(`TIP: Use 'cf target -o "%s" -s "%s"' to target new space`, orgName, spaceName))
					})

					It("attempts to make the user a space manager", func() {
						Expect(fakeActor.GrantSpaceManagerByUsernameCallCount()).To(Equal(1))
						orgGUID, spaceGUID, usernameArg := fakeActor.GrantSpaceManagerByUsernameArgsForCall(0)
						Expect(orgGUID).To(Equal("some-org-guid"))
						Expect(spaceGUID).To(Equal("some-space-guid"))
						Expect(usernameArg).To(Equal(username))
					})

					When("making the user a space manager succeeds", func() {
						BeforeEach(func() {
							fakeActor.GrantSpaceManagerByUsernameReturns(
								v2action.Warnings{"space-manager-warning-1", "space-manager-warning-2"},
								nil,
							)
						})

						It("prints the warnings", func() {
							Expect(executeErr).ToNot(HaveOccurred())
							Expect(testUI.Err).To(Say("space-manager-warning-1\nspace-manager-warning-2\n"))
						})

						It("attempts to make the user a space developer", func() {
							Expect(fakeActor.GrantSpaceDeveloperByUsernameCallCount()).To(Equal(1))
							spaceGUID, usernameArg := fakeActor.GrantSpaceDeveloperByUsernameArgsForCall(0)
							Expect(spaceGUID).To(Equal("some-space-guid"))
							Expect(usernameArg).To(Equal(username))
						})

						When("making the user a space developer succeeds", func() {
							BeforeEach(func() {
								fakeActor.GrantSpaceDeveloperByUsernameReturns(
									v2action.Warnings{"space-developer-warning", "other-warning"},
									nil,
								)
							})

							It("prints the warnings", func() {
								Expect(testUI.Err).To(Say("space-developer-warning"))
								Expect(testUI.Err).To(Say("other-warning"))
							})
						})

						When("making the user a space developer fails", func() {
							BeforeEach(func() {
								fakeActor.GrantSpaceDeveloperByUsernameReturns(
									v2action.Warnings{"space-developer-warning", "other-warning"},
									errors.New("Some terrible failure case"),
								)
							})

							It("returns the error", func() {
								Expect(executeErr).To(MatchError("Some terrible failure case"))
							})

							It("prints the warnings", func() {
								Expect(testUI.Err).To(Say("space-developer-warning"))
								Expect(testUI.Err).To(Say("other-warning"))
							})
						})
					})

					When("making the user a space manager fails", func() {
						BeforeEach(func() {
							fakeActor.GrantSpaceManagerByUsernameReturns(
								v2action.Warnings{"space-manager-warning-1", "space-manager-warning-2"},
								errors.New("some error"),
							)
						})

						It("should print all the warnings and returns the error", func() {
							Expect(executeErr).To(MatchError("some error"))
							Expect(testUI.Out).To(Say(`Assigning role SpaceManager to user %s in org %s / space %s as %s\.\.\.`, username, orgName, spaceName, username))
							Expect(testUI.Err).To(Say("space-manager-warning-1\nspace-manager-warning-2\n"))
						})

						It("should not try to grant the user SpaceDeveloper", func() {
							Expect(fakeActor.GrantSpaceDeveloperByUsernameCallCount()).To(Equal(0))
						})
					})
				})

				When("creating the space fails", func() {
					BeforeEach(func() {
						fakeActor.CreateSpaceReturns(
							v2action.Space{},
							v2action.Warnings{"some warning"},
							errors.New("some error"),
						)
					})

					It("should print the warnings and return the error", func() {
						Expect(executeErr).To(MatchError("some error"))
						Expect(testUI.Err).To(Say("some warning\n"))
					})
				})

				When("quota is not specified", func() {
					BeforeEach(func() {
						cmd.Quota = ""
					})
					It("attempts to create the space with no quota", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						_, _, inputQuota := fakeActor.CreateSpaceArgsForCall(0)
						Expect(inputQuota).To(BeEmpty())
					})

				})

				When("quota is specified", func() {
					BeforeEach(func() {
						cmd.Quota = "some-quota"
					})

					It("attempts to create the space with no quota", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						_, _, inputQuota := fakeActor.CreateSpaceArgsForCall(0)
						Expect(inputQuota).To(Equal("some-quota"))
					})
				})
			})
		})

		When("the server returns an already exists error", func() {
			BeforeEach(func() {
				fakeActor.CreateSpaceReturns(v2action.Space{}, v2action.Warnings{"already-exists-warnings"}, actionerror.SpaceNameTakenError{})
			})

			It("displays a warning but does not error", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(testUI.Err).To(Say("already-exists-warnings"))
				Expect(testUI.Out).To(Say("OK\n"))
				Expect(testUI.Err).To(Say("Space %s already exists", spaceName))
			})
		})
	})
})
