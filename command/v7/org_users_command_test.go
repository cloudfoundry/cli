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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("org-users Command", func() {
	var (
		cmd             OrgUsersCommand
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

		cmd = OrgUsersCommand{
			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
			AllUsers: false,
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
		})

		When("getting the current user fails", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserReturns(configv3.User{}, errors.New("get-current-user-error"))
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError("get-current-user-error"))
			})
		})

		When("getting the current user succeeds", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserReturns(
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
						billingManager := resources.User{
							Origin:           "uaa",
							PresentationName: "billing-manager",
							GUID:             "billingManager-guid",
						}
						orgAuditor := resources.User{
							Origin:           "uaa",
							PresentationName: "org-auditor",
							GUID:             "orgAuditor-guid",
						}
						orgUser := resources.User{
							Origin:           "uaa",
							PresentationName: "org-user",
							GUID:             "orgUser-guid",
						}

						orgUsersByRole := map[constant.RoleType][]resources.User{
							constant.OrgManagerRole:        {uaaAdmin, ldapAdmin, abbyUser, client},
							constant.OrgBillingManagerRole: {billingManager, uaaAdmin},
							constant.OrgAuditorRole:        {orgAuditor},
							constant.OrgUserRole:           {orgUser},
						}

						fakeActor.GetOrgUsersByRoleTypeReturns(
							orgUsersByRole,
							v7action.Warnings{"get-org-by-name-warning"},
							nil)
					})

					It("displays the alphabetized org-users in the org with origins", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say(`Getting users in org some-org-name as some-user\.\.\.`))
						Expect(testUI.Out).To(Say(`\n`))
						Expect(testUI.Out).To(Say(`\nORG MANAGER`))
						Expect(testUI.Out).To(Say(`\n  abby \(ldap\)`))
						Expect(testUI.Out).To(Say(`\n  admin \(uaa\)`))
						Expect(testUI.Out).To(Say(`\n  admin \(ldap\)`))
						Expect(testUI.Out).To(Say(`\n  admin \(client\)`))
						Expect(testUI.Out).To(Say(`\n`))
						Expect(testUI.Out).To(Say(`\nBILLING MANAGER`))
						Expect(testUI.Out).To(Say(`\n  billing-manager \(uaa\)`))
						Expect(testUI.Out).To(Say(`\n`))
						Expect(testUI.Out).To(Say(`\nORG AUDITOR`))
						Expect(testUI.Out).To(Say(`\n  org-auditor \(uaa\)`))

						Expect(testUI.Err).To(Say("get-org-by-name-warning"))

						Expect(fakeActor.GetOrganizationByNameCallCount()).To(Equal(1))
					})

					When("the --all-users flag is passed in", func() {
						BeforeEach(func() {
							cmd.AllUsers = true
						})

						It("displays the alphabetized org-users in the org with origins", func() {
							Expect(executeErr).ToNot(HaveOccurred())

							Expect(testUI.Out).To(Say(`Getting users in org some-org-name as some-user\.\.\.`))
							Expect(testUI.Out).To(Say("\n\n"))
							Expect(testUI.Out).To(Say("USERS"))
							Expect(testUI.Out).To(Say(`abby \(ldap\)`))
							Expect(testUI.Out).To(Say(`admin \(uaa\)`))
							// Ensure that admin (uaa) does not appear twice, even though it has two roles
							Expect(testUI.Out).NotTo(Say(`admin \(uaa\)`))
							Expect(testUI.Out).To(Say(`admin \(ldap\)`))
							Expect(testUI.Out).To(Say(`admin \(client\)`))
							Expect(testUI.Out).To(Say(`billing-manager \(uaa\)`))
							Expect(testUI.Out).To(Say(`org-auditor \(uaa\)`))
							Expect(testUI.Out).To(Say(`org-user \(uaa\)`))

							Expect(testUI.Err).To(Say("get-org-by-name-warning"))

							Expect(fakeActor.GetOrganizationByNameCallCount()).To(Equal(1))
						})
					})
				})

				When("There are no org users", func() {
					BeforeEach(func() {
						orgUsersByRole := map[constant.RoleType][]resources.User{
							constant.OrgManagerRole:        {},
							constant.OrgBillingManagerRole: {},
							constant.OrgAuditorRole:        {},
						}

						fakeActor.GetOrgUsersByRoleTypeReturns(
							orgUsersByRole,
							v7action.Warnings{"get-org-users-warning"},
							nil)
					})

					It("displays the headings with an informative 'not found' message", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say(`Getting users in org some-org-name as some-user\.\.\.`))
						Expect(testUI.Out).To(Say("\n\n"))
						Expect(testUI.Out).To(Say("ORG MANAGER"))
						Expect(testUI.Out).To(Say("No ORG MANAGER found"))
						Expect(testUI.Out).To(Say("\n\n"))
						Expect(testUI.Out).To(Say("BILLING MANAGER"))
						Expect(testUI.Out).To(Say("No BILLING MANAGER found"))
						Expect(testUI.Out).To(Say("\n\n"))
						Expect(testUI.Out).To(Say("ORG AUDITOR"))
						Expect(testUI.Out).To(Say("No ORG AUDITOR found"))

						Expect(testUI.Err).To(Say("get-org-users-warning"))

						Expect(fakeActor.GetOrganizationByNameCallCount()).To(Equal(1))
					})
				})

				When("there is an error getting org-users", func() {
					BeforeEach(func() {
						fakeActor.GetOrgUsersByRoleTypeReturns(
							nil,
							v7action.Warnings{"get-org-users-warning"},
							errors.New("get-org-users-error"))
					})

					It("returns an error with warnings", func() {
						Expect(executeErr).To(MatchError("get-org-users-error"))
						Expect(testUI.Err).To(Say("get-org-users-warning"))
					})
				})
			})
		})
	})
})
