package v7_test

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	v7 "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("create-org-quota Command", func() {
	var (
		cmd             v7.CreateOrgQuotaCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeCreateOrgQuotaActor
		orgQuotaName    string
		executeErr      error

		currentUserName string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeCreateOrgQuotaActor)

		cmd = v7.CreateOrgQuotaCommand{}

		currentUserName = "bob"
		fakeConfig.CurrentUserReturns(configv3.User{Name: currentUserName}, nil)
	})

	JustBeforeEach(func() {
		cmd = v7.CreateOrgQuotaCommand{
			UI:           testUI,
			Config:       fakeConfig,
			SharedActor:  fakeSharedActor,
			Actor:        fakeActor,
			RequiredArgs: flag.OrganizationQuota{OrganizationQuota: orgQuotaName},
		}
		executeErr = cmd.Execute(nil)
	})

	When("the environment is not set up correctly", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeFalse())
			Expect(checkTargetedSpace).To(BeFalse())
		})
	})

	When("the organization quota already exists", func() {
		BeforeEach(func() {
			fakeActor.CreateOrganizationQuotaReturns(v7action.Warnings{"warn-abc"}, ccerror.OrgQuotaAlreadyExists{"quota-already-exists"})
		})

		It("displays that it already exists, but does not return an error", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(fakeActor.CreateOrganizationQuotaCallCount()).To(Equal(1))
			Expect(testUI.Err).To(Say("warn-abc"))
			Expect(testUI.Out).To(Say("quota-already-exists"))
		})
	})

	When("creating the organization quota fails", func() {
		BeforeEach(func() {
			fakeActor.CreateOrganizationQuotaReturns(v7action.Warnings{"warn-456", "warn-789"}, errors.New("create-org-quota-err"))
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError("create-org-quota-err"))

			Expect(fakeActor.CreateOrganizationQuotaCallCount()).To(Equal(1))
			Expect(testUI.Err).To(Say("warn-456"))
			Expect(testUI.Err).To(Say("warn-789"))
		})
	})

	When("the org quota is created successfully", func() {
		// before each to make sure the env is set up
		BeforeEach(func() {
			orgQuotaName = "new-org-quota"
			fakeActor.CreateOrganizationQuotaReturns(
				v7action.Warnings{"warning"},
				nil)
		})
		It("creates a quota with a given name", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(fakeActor.CreateOrganizationQuotaCallCount()).To(Equal(1))

			quota := fakeActor.CreateOrganizationQuotaArgsForCall(0)
			Expect(quota).To(Equal(orgQuotaName))
			Expect(testUI.Out).To(Say("Creating org quota new-org-quota as bob..."))
			Expect(testUI.Out).To(Say("OK"))
		})
	})

})
