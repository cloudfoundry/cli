package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("space-users Command", func() {
	var (
		cmd             SpaceUsersCommand
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

		cmd = SpaceUsersCommand{
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

	When("an error is encountered checking if the environment is setup correctly", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: binaryName}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrgArg, checkTargetedSpaceArg := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrgArg).To(BeFalse())
			Expect(checkTargetedSpaceArg).To(BeFalse())

		})
	})

	When("the user is logged in and an org is targeted", func() {
		BeforeEach(func() {
			cmd.RequiredArgs.Organization = "some-org-name"
			cmd.RequiredArgs.Space = "some-space-name"
		})

		When("getting the current user fails", func() {
			BeforeEach(func() {
				fakeActor.GetCurrentUserReturns(configv3.User{}, errors.New("get-current-user-error"))
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError("get-current-user-error"))
			})
		})

		When("getting the current user succeeds", func() {
			BeforeEach(func() {
				fakeActor.GetCurrentUserReturns(
					configv3.User{Name: "some-user"},
					nil)
			})

			When("getting the org guid fails", func() {
				BeforeEach(func() {
					fakeActor.GetOrganizationByNameReturns(
						resources.Organization{},
						v7action.Warnings{"get-org-by-name-warning"},
						errors.New("get-org-by-name-error"))
				})

				It("returns the error and warnings", func() {
					Expect(executeErr).To(MatchError("get-org-by-name-error"))
					Expect(testUI.Err).To(Say("get-org-by-name-warning"))
				})
			})

			When("getting the org guid succeeds", func() {
				BeforeEach(func() {
					fakeActor.GetOrganizationByNameReturns(
						resources.Organization{
							Name: "org-1",
							GUID: "org-guid",
						},
						v7action.Warnings{"get-org-by-name-warning"},
						nil)
				})

				When("getting the space guid fails", func() {
					BeforeEach(func() {
						fakeActor.GetSpaceByNameAndOrganizationReturns(
							resources.Space{},
							v7action.Warnings{"get-space-by-name-and-org-warning"},
							errors.New("get-space-by-name-and-org-error"))
					})

					It("returns the error and warnings", func() {
						Expect(executeErr).To(MatchError("get-space-by-name-and-org-error"))
						Expect(testUI.Err).To(Say("get-space-by-name-and-org-warning"))
					})
				})

				When("getting the space guid succeeds", func() {
					BeforeEach(func() {
						fakeActor.GetOrganizationByNameReturns(
							resources.Organization{
								Name: "org-1",
								GUID: "org-guid",
							},
							v7action.Warnings{"get-org-by-name-warning"},
							nil)
					})

					When("There are all types of users", func() {
						BeforeEach(func() {
							abbyUser := resources.User{
								Origin:           "ldap",
								PresentationName: "abby",
								GUID:             "abby-user-guid",
							}
							uaaAdmin := resources.User{
								Origin:           "uaa",
								PresentationName: "admin",
								GUID:             "uaaAdmin-guid",
							}
							ldapAdmin := resources.User{
								Origin:           "ldap",
								PresentationName: "admin",
								GUID:             "ldapAdmin-guid",
							}
							client := resources.User{
								Origin:           "",
								PresentationName: "admin",
								GUID:             "client-guid",
							}
							spaceDeveloper := resources.User{
								Origin:           "uaa",
								PresentationName: "billing-manager",
								GUID:             "spaceDeveloper-guid",
							}
							spaceSupporter := resources.User{
								Origin:           "uaa",
								PresentationName: "fred",
								GUID:             "spaceSupporter-guid",
							}
							spaceAuditor := resources.User{
								Origin:           "uaa",
								PresentationName: "org-auditor",
								GUID:             "spaceAuditor-guid",
							}

							spaceUsersByRole := map[constant.RoleType][]resources.User{
								constant.SpaceManagerRole:   {uaaAdmin, ldapAdmin, abbyUser, client},
								constant.SpaceDeveloperRole: {spaceDeveloper},
								constant.SpaceSupporterRole: {spaceSupporter},
								constant.SpaceAuditorRole:   {spaceAuditor},
							}

							fakeActor.GetSpaceUsersByRoleTypeReturns(
								spaceUsersByRole,
								v7action.Warnings{"get-space-users-by-name-warning"},
								nil)
						})

						It("displays the alphabetized space-users in the space with origins", func() {
							Expect(executeErr).ToNot(HaveOccurred())

							Expect(testUI.Out).To(Say("Getting users in org some-org-name / space some-space-name as some-user..."))
							Expect(testUI.Out).To(Say(`\n`))
							Expect(testUI.Out).To(Say(`\nSPACE MANAGER`))
							Expect(testUI.Out).To(Say(`\n  abby \(ldap\)`))
							Expect(testUI.Out).To(Say(`\n  admin \(uaa\)`))
							Expect(testUI.Out).To(Say(`\n  admin \(ldap\)`))
							Expect(testUI.Out).To(Say(`\n  admin \(client\)`))
							Expect(testUI.Out).To(Say(`\n`))
							Expect(testUI.Out).To(Say(`\nSPACE DEVELOPER`))
							Expect(testUI.Out).To(Say(`\n  billing-manager \(uaa\)`))
							Expect(testUI.Out).To(Say(`\n`))
							Expect(testUI.Out).To(Say(`\nSPACE SUPPORTER`))
							Expect(testUI.Out).To(Say(`\n  fred \(uaa\)`))
							Expect(testUI.Out).To(Say(`\n`))
							Expect(testUI.Out).To(Say(`\nSPACE AUDITOR`))
							Expect(testUI.Out).To(Say(`\n  org-auditor \(uaa\)`))

							Expect(testUI.Err).To(Say("get-space-users-by-name-warning"))

							Expect(fakeActor.GetOrganizationByNameCallCount()).To(Equal(1))
						})
					})

					When("There are no space users", func() {
						BeforeEach(func() {
							spaceUsersByRole := map[constant.RoleType][]resources.User{
								constant.OrgManagerRole:        {},
								constant.OrgBillingManagerRole: {},
								constant.OrgAuditorRole:        {},
							}

							fakeActor.GetSpaceUsersByRoleTypeReturns(
								spaceUsersByRole,
								v7action.Warnings{"get-space-users-warning"},
								nil)
						})

						It("displays the headings with an informative 'not found' message", func() {
							Expect(executeErr).ToNot(HaveOccurred())

							Expect(testUI.Out).To(Say("Getting users in org some-org-name / space some-space-name as some-user..."))
							Expect(testUI.Out).To(Say("\n\n"))
							Expect(testUI.Out).To(Say(`SPACE MANAGER`))
							Expect(testUI.Out).To(Say("No SPACE MANAGER found"))
							Expect(testUI.Out).To(Say("\n\n"))
							Expect(testUI.Out).To(Say(`SPACE DEVELOPER`))
							Expect(testUI.Out).To(Say("No SPACE DEVELOPER found"))
							Expect(testUI.Out).To(Say("\n\n"))
							Expect(testUI.Out).To(Say(`SPACE AUDITOR`))
							Expect(testUI.Out).To(Say("No SPACE AUDITOR found"))

							Expect(testUI.Err).To(Say("get-space-users-warning"))

							Expect(fakeActor.GetOrganizationByNameCallCount()).To(Equal(1))
						})
					})

					When("there is an error getting space-users", func() {
						BeforeEach(func() {
							fakeActor.GetSpaceUsersByRoleTypeReturns(
								nil,
								v7action.Warnings{"get-space-users-warning"},
								errors.New("get-space-users-error"))
						})

						It("returns an error with warnings", func() {
							Expect(executeErr).To(MatchError("get-space-users-error"))
							Expect(testUI.Err).To(Say("get-space-users-warning"))
						})
					})
				})
			})
		})
	})
})
