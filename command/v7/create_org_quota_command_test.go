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
	"code.cloudfoundry.org/cli/types"
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

	When("the org quota includes invalid flag values", func() {
		// before each to make sure the env is set up
		When("the -a flag has an invalid value", func() {
			BeforeEach(func() {
				cmd.NumAppInstances = "hello"
			})
			It("returns an error ", func() {
				Expect(executeErr).To(HaveOccurred())
				Expect(executeErr).To(MatchError("Invalid value for flag -a: hello"))
			})
		})

		When("the -i flag has an invalid value", func() {
			BeforeEach(func() {
				cmd.PerProcessMemory = "hello"
			})
			It("returns an error ", func() {
				Expect(executeErr).To(HaveOccurred())
				Expect(executeErr).To(MatchError("Invalid value for flag -i: hello"))
			})
		})

		When("the -m flag has an invalid value", func() {
			BeforeEach(func() {
				cmd.TotalMemory = "hello"
			})
			It("returns an error ", func() {
				Expect(executeErr).To(HaveOccurred())
				Expect(executeErr).To(MatchError("Invalid value for flag -m: hello"))
			})
		})

		When("the -r flag has an invalid value", func() {
			BeforeEach(func() {
				cmd.TotalRoutes = "hello"
			})
			It("returns an error ", func() {
				Expect(executeErr).To(HaveOccurred())
				Expect(executeErr).To(MatchError("Invalid value for flag -r: hello"))
			})
		})

		When("the --reserved-route-ports flag has an invalid value", func() {
			BeforeEach(func() {
				cmd.TotalReservedPorts = "hello"
			})
			It("returns an error ", func() {
				Expect(executeErr).To(HaveOccurred())
				Expect(executeErr).To(MatchError("Invalid value for flag --reserved-route-ports: hello"))
			})
		})

		When("the -s flag has an invalid value", func() {
			BeforeEach(func() {
				cmd.TotalServiceInstances = "hello"
			})
			It("returns an error ", func() {
				Expect(executeErr).To(HaveOccurred())
				Expect(executeErr).To(MatchError("Invalid value for flag -s: hello"))
			})
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
		When("using the default values", func() {
			BeforeEach(func() {
				fakeActor.CreateOrganizationQuotaReturns(
					v7action.Warnings{"warning"},
					nil)
			})

			It("creates a quota with a given name", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeActor.CreateOrganizationQuotaCallCount()).To(Equal(1))

				quota := fakeActor.CreateOrganizationQuotaArgsForCall(0)
				Expect(quota.Name).To(Equal(orgQuotaName))
				Expect(quota.PaidServicePlans).To(BeFalse())
				Expect(quota.TotalServiceInstances.Value).To(Equal(0))
				Expect(quota.TotalServiceInstances.IsSet).To(BeTrue())

				Expect(quota.TotalRoutes.Value).To(Equal(0))
				Expect(quota.TotalRoutes.IsSet).To(BeTrue())
				Expect(quota.TotalReservedPorts.Value).To(Equal(0))
				Expect(quota.TotalReservedPorts.IsSet).To(BeTrue())

				Expect(quota.TotalAppInstances.IsSet).To(BeFalse())
				Expect(quota.InstanceMemory.IsSet).To(BeFalse())
				Expect(quota.TotalMemory.IsSet).To(BeFalse())

				Expect(testUI.Out).To(Say("Creating org quota %s as bob...", orgQuotaName))
				Expect(testUI.Out).To(Say("OK"))
			})
		})

		When("the org quota is provided a value for each flag", func() {
			// before each to make sure the env is set up
			BeforeEach(func() {
				cmd.PaidServicePlans = true
				cmd.NumAppInstances = "10"
				cmd.PerProcessMemory = flag.Megabytes{types.NullUint64{IsSet: true, Value: 9 }}
				cmd.TotalMemory = flag.Megabytes{types.NullUint64{IsSet: true, Value: 2048 }}
				cmd.TotalRoutes = "7"
				cmd.TotalReservedPorts = "1"
				cmd.TotalServiceInstances = "2"
				fakeActor.CreateOrganizationQuotaReturns(
					v7action.Warnings{"warning"},
					nil)
			})
			It("creates a quota with the values given from the flags", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeActor.CreateOrganizationQuotaCallCount()).To(Equal(1))

				quota := fakeActor.CreateOrganizationQuotaArgsForCall(0)
				Expect(quota.Name).To(Equal(orgQuotaName))

				Expect(quota.TotalAppInstances.IsSet).To(Equal(true))
				Expect(quota.TotalAppInstances.Value).To(Equal(10))

				Expect(quota.InstanceMemory.IsSet).To(Equal(true))
				Expect(quota.InstanceMemory.Value).To(Equal(9))

				Expect(quota.PaidServicePlans).To(Equal(true))

				Expect(quota.TotalMemory.IsSet).To(Equal(true))
				Expect(quota.TotalMemory.Value).To(Equal(2048))

				Expect(quota.TotalRoutes.IsSet).To(Equal(true))
				Expect(quota.TotalRoutes.Value).To(Equal(7))

				Expect(quota.TotalReservedPorts.IsSet).To(Equal(true))
				Expect(quota.TotalReservedPorts.Value).To(Equal(1))

				Expect(quota.TotalServiceInstances.IsSet).To(Equal(true))
				Expect(quota.TotalServiceInstances.Value).To(Equal(2))

				Expect(testUI.Out).To(Say("Creating org quota %s as bob...", orgQuotaName))
				Expect(testUI.Out).To(Say("OK"))
			})
		})

		When("every value is set to unlimited", func() {
			// before each to make sure the env is set up
			BeforeEach(func() {
				cmd.PaidServicePlans = true
				cmd.NumAppInstances = flag.Instances{types.NullInt{IsSet: true, Value: 10 }}
				cmd.PerProcessMemory = flag.Megabytes{types.NullUint64{IsSet: true, Value: 9 }}
				cmd.TotalMemory = flag.Megabytes{types.NullUint64{IsSet: true, Value: 2048 }}
				cmd.TotalRoutes = types.NullInt{IsSet: true , Value: 7}
				cmd.TotalReservedPorts = types.NullInt{IsSet: true , Value: 1}
				cmd.TotalServiceInstances = flag.Instances{types.NullInt{IsSet: true, Value: 2 }}
				fakeActor.CreateOrganizationQuotaReturns(
					v7action.Warnings{"warning"},
					nil)
			})
			It("creates a quota with the values given from the flags", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeActor.CreateOrganizationQuotaCallCount()).To(Equal(1))

				quota := fakeActor.CreateOrganizationQuotaArgsForCall(0)
				Expect(quota.Name).To(Equal(orgQuotaName))

				Expect(quota.TotalAppInstances.IsSet).To(Equal(true))
				Expect(quota.TotalAppInstances.Value).To(Equal(10))

				Expect(quota.InstanceMemory.IsSet).To(Equal(true))
				Expect(quota.InstanceMemory.Value).To(Equal(9))

				Expect(quota.PaidServicePlans).To(Equal(true))

				Expect(quota.TotalMemory.IsSet).To(Equal(true))
				Expect(quota.TotalMemory.Value).To(Equal(2048))

				Expect(quota.TotalRoutes.IsSet).To(Equal(true))
				Expect(quota.TotalRoutes.Value).To(Equal(7))

				Expect(quota.TotalReservedPorts.IsSet).To(Equal(true))
				Expect(quota.TotalReservedPorts.Value).To(Equal(1))

				Expect(quota.TotalServiceInstances.IsSet).To(Equal(true))
				Expect(quota.TotalServiceInstances.Value).To(Equal(2))

				Expect(testUI.Out).To(Say("Creating org quota %s as bob...", orgQuotaName))
				Expect(testUI.Out).To(Say("OK"))
			})
		})
	})
})
