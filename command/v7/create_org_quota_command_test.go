package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	v7 "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
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
		orgQuotaName = "new-org-quota-name"

		cmd = v7.CreateOrgQuotaCommand{
			UI:           testUI,
			Config:       fakeConfig,
			SharedActor:  fakeSharedActor,
			Actor:        fakeActor,
			RequiredArgs: flag.OrganizationQuota{OrganizationQuotaName: orgQuotaName},
		}

		currentUserName = "bob"
		fakeConfig.CurrentUserReturns(configv3.User{Name: currentUserName}, nil)
	})

	JustBeforeEach(func() {

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
			fakeActor.CreateOrganizationQuotaReturns(v7action.Warnings{"warn-abc"}, ccerror.QuotaAlreadyExists{"quota-already-exists"})
		})

		It("displays that it already exists, but does not return an error", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(fakeActor.CreateOrganizationQuotaCallCount()).To(Equal(1))
			Expect(testUI.Err).To(Say("warn-abc"))
			Expect(testUI.Err).To(Say("quota-already-exists"))
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
		When("the org quota is provided a value for each flag", func() {
			// before each to make sure the env is set up
			BeforeEach(func() {
				cmd.PaidServicePlans = true
				cmd.NumAppInstances = flag.IntegerLimit{IsSet: true, Value: 10}
				cmd.PerProcessMemory = flag.MemoryWithUnlimited{IsSet: true, Value: 9}
				cmd.TotalMemory = flag.MemoryWithUnlimited{IsSet: true, Value: 2048}
				cmd.TotalRoutes = flag.IntegerLimit{IsSet: true, Value: 7}
				cmd.TotalReservedPorts = flag.IntegerLimit{IsSet: true, Value: 1}
				cmd.TotalServiceInstances = flag.IntegerLimit{IsSet: true, Value: 2}
				fakeActor.CreateOrganizationQuotaReturns(
					v7action.Warnings{"warning"},
					nil)
			})

			It("creates a quota with the values given from the flags", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeActor.CreateOrganizationQuotaCallCount()).To(Equal(1))

				quotaName, quotaLimits := fakeActor.CreateOrganizationQuotaArgsForCall(0)
				Expect(quotaName).To(Equal(orgQuotaName))

				Expect(quotaLimits.TotalInstances.IsSet).To(Equal(true))
				Expect(quotaLimits.TotalInstances.Value).To(Equal(10))

				Expect(quotaLimits.PerProcessMemoryInMB.IsSet).To(Equal(true))
				Expect(quotaLimits.PerProcessMemoryInMB.Value).To(Equal(9))

				Expect(quotaLimits.PaidServicesAllowed).To(Equal(true))

				Expect(quotaLimits.TotalMemoryInMB.IsSet).To(Equal(true))
				Expect(quotaLimits.TotalMemoryInMB.Value).To(Equal(2048))

				Expect(quotaLimits.TotalRoutes.IsSet).To(Equal(true))
				Expect(quotaLimits.TotalRoutes.Value).To(Equal(7))

				Expect(quotaLimits.TotalReservedPorts.IsSet).To(Equal(true))
				Expect(quotaLimits.TotalReservedPorts.Value).To(Equal(1))

				Expect(quotaLimits.TotalServiceInstances.IsSet).To(Equal(true))
				Expect(quotaLimits.TotalServiceInstances.Value).To(Equal(2))

				Expect(testUI.Out).To(Say("Creating org quota %s as bob...", orgQuotaName))
				Expect(testUI.Out).To(Say("OK"))
			})
		})
	})
})
