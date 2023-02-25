package v7_test

import (
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("services command", func() {
	const (
		org       = "fake-org"
		space     = "fake-space"
		spaceGUID = "fake-space-guid"
		username  = "fake-user"
	)

	var (
		cmd             ServicesCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		cmd = ServicesCommand{
			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		fakeConfig.TargetedSpaceReturns(configv3.Space{
			GUID: spaceGUID,
			Name: space,
		})

		fakeConfig.TargetedOrganizationReturns(configv3.Organization{
			Name: org,
		})

		fakeActor.GetCurrentUserReturns(configv3.User{Name: username}, nil)

		fakeActor.GetServiceInstancesForSpaceReturns(
			[]v7action.ServiceInstance{
				{
					Name:                "msi1",
					Type:                resources.ManagedServiceInstance,
					ServicePlanName:     "fake-plan-1",
					ServiceOfferingName: "fake-offering-1",
					ServiceBrokerName:   "fake-broker-1",
					UpgradeAvailable:    types.NewOptionalBoolean(true),
					BoundApps:           []string{"foo", "bar"},
					LastOperation:       "create succeeded",
				},
				{
					Name:                "msi2",
					Type:                resources.ManagedServiceInstance,
					ServicePlanName:     "fake-plan-2",
					ServiceOfferingName: "fake-offering-2",
					ServiceBrokerName:   "fake-broker-2",
					UpgradeAvailable:    types.NewOptionalBoolean(false),
					BoundApps:           []string{"baz", "quz"},
					LastOperation:       "delete in progress",
				},
				{
					Name:                "msi3",
					Type:                resources.ManagedServiceInstance,
					ServicePlanName:     "fake-plan-3",
					ServiceOfferingName: "fake-offering-3",
					ServiceBrokerName:   "fake-broker-2",
					BoundApps:           []string{},
					LastOperation:       "update failed",
				},
				{
					Name:      "upsi1",
					Type:      resources.UserProvidedServiceInstance,
					BoundApps: []string{"foo", "bar"},
				},
				{
					Name:      "upsi2",
					Type:      resources.UserProvidedServiceInstance,
					BoundApps: []string{"baz", "qux"},
				},
				{
					Name:      "upsi3",
					Type:      resources.UserProvidedServiceInstance,
					BoundApps: []string{},
				},
			},
			v7action.Warnings{"something silly"},
			nil,
		)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	It("checks that the user is logged in and a space is targeted", func() {
		Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
		actualSpace, actualOrg := fakeSharedActor.CheckTargetArgsForCall(0)
		Expect(actualSpace).To(BeTrue())
		Expect(actualOrg).To(BeTrue())
	})

	It("prints an introductory message", func() {
		Expect(testUI.Out).To(Say(`Getting service instances in org %s / space %s as %s...\n\n`, org, space, username))
	})

	It("asks the actor to get the service instances", func() {
		Expect(fakeActor.GetServiceInstancesForSpaceCallCount()).To(Equal(1))
		Expect(fakeActor.GetServiceInstancesForSpaceArgsForCall(0)).To(Equal(spaceGUID))
	})

	It("prints a table with the services, and warning", func() {
		Expect(executeErr).NotTo(HaveOccurred())
		Expect(testUI.Err).To(Say("something silly"))
		Expect(testUI.Out).To(SatisfyAll(
			Say(`name\s+offering\s+plan\s+bound apps\s+last operation\s+broker\s+upgrade available\n`),
			Say(`msi1\s+fake-offering-1\s+fake-plan-1\s+foo, bar\s+create succeeded\s+fake-broker-1\s+yes\n`),
			Say(`msi2\s+fake-offering-2\s+fake-plan-2\s+baz, quz\s+delete in progress\s+fake-broker-2\s+no\n`),
			Say(`msi3\s+fake-offering-3\s+fake-plan-3\s+update failed\s+fake-broker-2\s*\n`),
			Say(`upsi1\s+user-provided\s+foo, bar\s*\n`),
			Say(`upsi2\s+user-provided\s+baz, qux\s*\n`),
			Say(`upsi3\s+user-provided\s*\n`),
		))
	})

	When("omit apps is set", func() {
		BeforeEach(func() {
			cmd.OmitApps = true
		})
		It("doesn't print the bound apps table", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(testUI.Err).To(Say("something silly"))
			Expect(testUI.Out).To(SatisfyAll(
				Say(`name\s+offering\s+plan\s+last operation\s+broker\s+upgrade available\n`),
				Say(`msi1\s+fake-offering-1\s+fake-plan-1\s+create succeeded\s+fake-broker-1\s+yes\n`),
				Say(`msi2\s+fake-offering-2\s+fake-plan-2\s+delete in progress\s+fake-broker-2\s+no\n`),
				Say(`msi3\s+fake-offering-3\s+fake-plan-3\s+update failed\s+fake-broker-2\s*\n`),
				Say(`upsi1\s+user-provided\s*\n`),
				Say(`upsi2\s+user-provided\s*\n`),
				Say(`upsi3\s+user-provided\s*\n`),
			))
		})
	})

	When("there are no service instances", func() {
		BeforeEach(func() {
			fakeActor.GetServiceInstancesForSpaceReturns(
				[]v7action.ServiceInstance{},
				v7action.Warnings{"foo warning"},
				nil,
			)
		})

		It("says that none were found", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(testUI.Err).To(Say("foo warning"))
			Expect(testUI.Out).To(Say(`No service instances found\.`))
		})
	})

	When("checking target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(errors.New("not logged in"))
		})

		It("fails", func() {
			Expect(executeErr).To(MatchError("not logged in"))
		})
	})

	When("getting the user fails", func() {
		BeforeEach(func() {
			fakeActor.GetCurrentUserReturns(configv3.User{}, errors.New("bang"))
		})

		It("fails", func() {
			Expect(executeErr).To(MatchError("bang"))
		})
	})

	When("getting the instances fails", func() {
		BeforeEach(func() {
			fakeActor.GetServiceInstancesForSpaceReturns(
				[]v7action.ServiceInstance{},
				v7action.Warnings{"a warning"},
				errors.New("a bad thing happened"),
			)
		})

		It("fails and prints warnings", func() {
			Expect(testUI.Err).To(Say(`a warning\n`))
			Expect(executeErr).To(MatchError("a bad thing happened"))
		})
	})
})
