package v7_test

import (
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/ui"

	. "code.cloudfoundry.org/cli/command/v7"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("set-org-quota Command", func() {
	var (
		cmd             SetOrgQuotaCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeSetOrgQuotaActor
		binaryName      string
		executeErr      error
		input           *Buffer

		getOrgWarning     string
		applyQuotaWarning string
		currentUser       string
	)

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeSetOrgQuotaActor)
		getOrgWarning = RandomString("get-org-warning")
		applyQuotaWarning = RandomString("apply-quota-warning")

		cmd = SetOrgQuotaCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)

		currentUser = "current-user"
		fakeConfig.CurrentUserNameReturns(currentUser, nil)

		fakeActor.GetOrganizationByNameReturns(
			v7action.Organization{GUID: "some-org-guid"},
			v7action.Warnings{getOrgWarning},
			nil,
		)

		fakeActor.ApplyOrganizationQuotaByNameReturns(
			v7action.Warnings{applyQuotaWarning},
			nil,
		)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("the environment is not set up correctly", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(errors.New("check-target-failure"))
		})

		It("does not require a targeted org or space", func() {
			targetedOrgRequired, targetedSpaceRequired := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(targetedOrgRequired).To(BeFalse())
			Expect(targetedSpaceRequired).To(BeFalse())
		})

		It("should return the error", func() {
			Expect(executeErr).To(MatchError("check-target-failure"))
		})
	})

	When("applying the quota succeeds", func() {
		BeforeEach(func() {
			cmd.RequiredArgs = flag.SetOrgQuotaArgs{}
			cmd.RequiredArgs.Organization = RandomString("org-name")
			cmd.RequiredArgs.OrganizationQuota = RandomString("org-quota-name")
		})

		It("gets the organization by name", func() {
			Expect(fakeActor.GetOrganizationByNameCallCount()).To(Equal(1))
		})

		It("applies the quota to the organization", func() {
			Expect(fakeActor.ApplyOrganizationQuotaByNameCallCount()).To(Equal(1))
			orgQuotaName, orgGUID := fakeActor.ApplyOrganizationQuotaByNameArgsForCall(0)
			Expect(orgQuotaName).To(Equal(cmd.RequiredArgs.OrganizationQuota))
			Expect(orgGUID).To(Equal("some-org-guid"))
		})

		It("displays warnings and returns without error", func() {
			Expect(executeErr).NotTo(HaveOccurred())

			Expect(testUI.Err).To(Say(getOrgWarning))
			Expect(testUI.Err).To(Say(applyQuotaWarning))
		})

		It("displays flavor text and returns without error", func() {
			Expect(executeErr).NotTo(HaveOccurred())

			Expect(testUI.Out).To(Say("Setting quota %s to org %s as %s...", cmd.RequiredArgs.OrganizationQuota, cmd.RequiredArgs.Organization, currentUser))
			Expect(testUI.Out).To(Say("OK"))
		})
	})

	When("getting the current user fails", func() {
		BeforeEach(func() {
			fakeConfig.CurrentUserNameReturns("", errors.New("current-user-error"))
		})

		It("returns the error", func() {
			Expect(executeErr).To(MatchError("current-user-error"))
		})
	})

	When("getting the org fails", func() {
		BeforeEach(func() {
			fakeActor.GetOrganizationByNameReturns(
				v7action.Organization{GUID: "some-org-guid"},
				v7action.Warnings{getOrgWarning},
				errors.New("get-org-error"),
			)
		})

		It("displays warnings and returns the error", func() {
			Expect(testUI.Err).To(Say(getOrgWarning))

			Expect(executeErr).To(MatchError("get-org-error"))
		})
	})

	When("applying the quota to the org fails", func() {
		BeforeEach(func() {
			fakeActor.ApplyOrganizationQuotaByNameReturns(
				v7action.Warnings{applyQuotaWarning},
				errors.New("apply-org-quota-error"),
			)
		})

		It("displays warnings and returns the error", func() {
			Expect(testUI.Err).To(Say(getOrgWarning))
			Expect(testUI.Err).To(Say(applyQuotaWarning))

			Expect(executeErr).To(MatchError("apply-org-quota-error"))
		})
	})
})
