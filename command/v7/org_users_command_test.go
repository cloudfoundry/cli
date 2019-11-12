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

var _ = Describe("org-users Command", func() {
	var (
		cmd             OrgUsersCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeRo
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeOrgsActor)

		cmd = OrgUsersCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
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
		When("getting the current user fails", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserReturns(configv3.User{}, errors.New("get-user-error"))
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError("get-user-error"))
			})
		})

		When("getting the current user succeeds", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserReturns(
					configv3.User{Name: "some-user"},
					nil)
			})

			BeforeEach(func() {
				fakeActor.GetOrganizationsReturns(
					[]v7action.Organization{
						{Name: "org-1"},
					},
					v7action.Warnings{"get-org-users-warning"},
					nil)
			})

			When("There are all types of users", func() {
				It("displays the alphabetized org-users in the org with origins", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say(`Getting users in org-1 as some-user\.\.\.`))
					Expect(testUI.Out).To(Say(""))
					Expect(testUI.Out).To(Say("ORG MANAGER"))
					Expect(testUI.Out).To(Say("abby (ldap)"))
					Expect(testUI.Out).To(Say("admin (UAA)"))
					Expect(testUI.Out).To(Say("admin (ldap)"))
					Expect(testUI.Out).To(Say("admin (client)"))
					Expect(testUI.Out).To(Say(""))
					Expect(testUI.Out).To(Say("BILLING MANAGER"))
					Expect(testUI.Out).To(Say("billing-mgr (UAA)"))
					Expect(testUI.Out).To(Say(""))
					Expect(testUI.Out).To(Say("ORG AUDITOR"))
					Expect(testUI.Out).To(Say("org-auditor-1 (UAA)"))

					Expect(testUI.Err).To(Say("get-org-users-warning"))

					Expect(fakeActor.GetOrganizationsCallCount()).To(Equal(1))
				})
			})

			When("There are no org users", func() {
				It("displays the headings with an informative 'not found' message", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say(`Getting users in org-1 as some-user\.\.\.`))
					Expect(testUI.Out).To(Say(""))
					Expect(testUI.Out).To(Say("ORG MANAGER"))
					Expect(testUI.Out).To(Say("No ORG MANAGER found"))
					Expect(testUI.Out).To(Say(""))
					Expect(testUI.Out).To(Say("BILLING MANAGER"))
					Expect(testUI.Out).To(Say("No BILLING MANAGER found"))
					Expect(testUI.Out).To(Say(""))
					Expect(testUI.Out).To(Say("ORG AUDITOR"))
					Expect(testUI.Out).To(Say("No ORG AUDITOR found"))

					Expect(testUI.Err).To(Say("get-org-users-warning"))

					Expect(fakeActor.GetOrganizationsCallCount()).To(Equal(1))
				})
			})

			When("a translatable error is encountered getting org-users", func() {
				BeforeEach(func() {
					fakeActor.GetOrganizationsReturns(
						nil,
						v7action.Warnings{"get-org-users-warning"},
						actionerror.OrganizationNotFoundError{Name: "not-found-org"})
				})

				It("returns a translatable error", func() {
					Expect(executeErr).To(MatchError(actionerror.OrganizationNotFoundError{Name: "not-found-org"}))

					Expect(testUI.Out).To(Say(`Getting org-users as some-user\.\.\.`))
					Expect(testUI.Out).To(Say(""))

					Expect(testUI.Err).To(Say("get-org-users-warning"))

					Expect(fakeActor.GetOrganizationsCallCount()).To(Equal(1))
				})
			})
		})
	})
})
