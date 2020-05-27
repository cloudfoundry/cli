package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/command/translatableerror"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
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

var _ = Describe("UpdateOrgQuotaCommand", func() {
	var (
		cmd             v7.UpdateOrgQuotaCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		orgQuotaName    string
		executeErr      error

		currentUserName string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)
		orgQuotaName = "old-org-quota-name"

		cmd = v7.UpdateOrgQuotaCommand{
			BaseCommand: v7.BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
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

	When("updating the organization quota fails", func() {
		BeforeEach(func() {
			fakeActor.UpdateOrganizationQuotaReturns(v7action.Warnings{"warn-456", "warn-789"}, errors.New("update-org-quota-err"))
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError("update-org-quota-err"))

			Expect(fakeActor.UpdateOrganizationQuotaCallCount()).To(Equal(1))
			Expect(testUI.Err).To(Say("warn-456"))
			Expect(testUI.Err).To(Say("warn-789"))
		})
	})

	When("the org quota is updated successfully", func() {
		When("the org quota is provided a new value for each flag", func() {
			BeforeEach(func() {
				cmd.NewName = "new-org-quota-name"
				cmd.PaidServicePlans = true
				cmd.NumAppInstances = flag.IntegerLimit{IsSet: true, Value: 10}
				cmd.PerProcessMemory = flag.MemoryWithUnlimited{IsSet: true, Value: 9}
				cmd.TotalMemory = flag.MemoryWithUnlimited{IsSet: true, Value: 2048}
				cmd.TotalRoutes = flag.IntegerLimit{IsSet: true, Value: 7}
				cmd.TotalReservedPorts = flag.IntegerLimit{IsSet: true, Value: 1}
				cmd.TotalServiceInstances = flag.IntegerLimit{IsSet: true, Value: 2}
				fakeActor.UpdateOrganizationQuotaReturns(
					v7action.Warnings{"warning"},
					nil)
			})

			It("updates the org quota to the new values", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(fakeActor.UpdateOrganizationQuotaCallCount()).To(Equal(1))

				oldQuotaName, newQuotaName, quotaLimits := fakeActor.UpdateOrganizationQuotaArgsForCall(0)

				Expect(oldQuotaName).To(Equal("old-org-quota-name"))
				Expect(newQuotaName).To(Equal("new-org-quota-name"))

				Expect(quotaLimits.TotalInstances.IsSet).To(Equal(true))
				Expect(quotaLimits.TotalInstances.Value).To(Equal(10))

				Expect(quotaLimits.PerProcessMemoryInMB.IsSet).To(Equal(true))
				Expect(quotaLimits.PerProcessMemoryInMB.Value).To(Equal(9))

				Expect(*quotaLimits.PaidServicesAllowed).To(Equal(true))

				Expect(quotaLimits.TotalMemoryInMB.IsSet).To(Equal(true))
				Expect(quotaLimits.TotalMemoryInMB.Value).To(Equal(2048))

				Expect(quotaLimits.TotalRoutes.IsSet).To(Equal(true))
				Expect(quotaLimits.TotalRoutes.Value).To(Equal(7))

				Expect(quotaLimits.TotalReservedPorts.IsSet).To(Equal(true))
				Expect(quotaLimits.TotalReservedPorts.Value).To(Equal(1))

				Expect(quotaLimits.TotalServiceInstances.IsSet).To(Equal(true))
				Expect(quotaLimits.TotalServiceInstances.Value).To(Equal(2))

				Expect(testUI.Out).To(Say("Updating org quota %s as bob...", orgQuotaName))
				Expect(testUI.Out).To(Say("OK"))
			})
		})

		When("only some org quota limits are updated", func() {
			BeforeEach(func() {
				cmd.TotalMemory = flag.MemoryWithUnlimited{IsSet: true, Value: 2048}
				cmd.TotalServiceInstances = flag.IntegerLimit{IsSet: true, Value: 2}
				fakeActor.UpdateOrganizationQuotaReturns(
					v7action.Warnings{"warning"},
					nil)
			})

			It("updates the org quota to the new values", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(fakeActor.UpdateOrganizationQuotaCallCount()).To(Equal(1))

				oldQuotaName, newQuotaName, quotaLimits := fakeActor.UpdateOrganizationQuotaArgsForCall(0)

				Expect(oldQuotaName).To(Equal("old-org-quota-name"))
				Expect(newQuotaName).To(Equal(""))

				Expect(quotaLimits.TotalServiceInstances.IsSet).To(Equal(true))
				Expect(quotaLimits.TotalServiceInstances.Value).To(Equal(2))

				Expect(quotaLimits.TotalMemoryInMB.IsSet).To(Equal(true))
				Expect(quotaLimits.TotalMemoryInMB.Value).To(Equal(2048))

				Expect(quotaLimits.TotalInstances).To(BeNil())

				Expect(quotaLimits.PerProcessMemoryInMB).To(BeNil())

				Expect(quotaLimits.PaidServicesAllowed).To(BeNil())

				Expect(quotaLimits.TotalRoutes).To(BeNil())

				Expect(quotaLimits.TotalReservedPorts).To(BeNil())

				Expect(testUI.Out).To(Say("Updating org quota %s as bob...", orgQuotaName))
				Expect(testUI.Out).To(Say("OK"))
			})
		})
	})

	When("conflicting flags are given", func() {
		BeforeEach(func() {
			cmd.PaidServicePlans = true
			cmd.NoPaidServicePlans = true
		})

		It("returns with a helpful error", func() {
			Expect(executeErr).To(MatchError(translatableerror.ArgumentCombinationError{
				Args: []string{"--allow-paid-service-plans", "--disallow-paid-service-plans"},
			}))
		})
	})
})
