package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("disable-service-access Command", func() {
	var (
		cmd             DisableServiceAccessCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		cmd = DisableServiceAccessCommand{
			BaseCommand: command.BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		setPositionalFlags(&cmd, "some-service")

		fakeActor.DisableServiceAccessReturns(nil, v7action.Warnings{"a warning"}, nil)
	})

	It("checks the target", func() {
		err := cmd.Execute(nil)
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
		org, space := fakeSharedActor.CheckTargetArgsForCall(0)
		Expect(org).To(BeFalse())
		Expect(space).To(BeFalse())
	})

	DescribeTable(
		"message text",
		func(plan, org, broker, expected string) {
			setPositionalFlags(&cmd, "fake-service")
			fakeConfig.CurrentUserReturns(configv3.User{Name: "fake-user"}, nil)

			setFlag(&cmd, "-o", org)
			setFlag(&cmd, "-p", plan)
			setFlag(&cmd, "-b", broker)

			err := cmd.Execute(nil)
			Expect(err).NotTo(HaveOccurred())

			Expect(testUI.Out).To(Say(expected))
		},
		Entry("no flags", "", "", "",
			`Disabling access to all plans of service fake-service for all orgs as fake-user\.\.\.`),
		Entry("plan", "fake-plan", "", "",
			`Disabling access to plan fake-plan of service fake-service for all orgs as fake-user\.\.\.`),
		Entry("org", "", "fake-org", "",
			`Disabling access to all plans of service fake-service for org fake-org as fake-user\.\.\.`),
		Entry("broker", "", "", "fake-broker",
			`Disabling access to all plans of service fake-service from broker fake-broker for all orgs as fake-user\.\.\.`),
		Entry("plan and org", "fake-plan", "fake-org", "",
			`Disabling access to plan fake-plan of service fake-service for org fake-org as fake-user\.\.\.`),
		Entry("plan and broker", "fake-plan", "", "fake-broker",
			`Disabling access to plan fake-plan of service fake-service from broker fake-broker for all orgs as fake-user\.\.\.`),
		Entry("plan, org and broker", "fake-plan", "fake-org", "fake-broker",
			`Disabling access to plan fake-plan of service fake-service from broker fake-broker for org fake-org as fake-user\.\.\.`),
		Entry("broker and org", "", "fake-org", "fake-broker",
			`Disabling access to all plans of service fake-service from broker fake-broker for org fake-org as fake-user\.\.\.`),
	)

	It("passes the right parameters to the actor", func() {
		const (
			offering = "myoffering"
			broker   = "mybroker"
			plan     = "myplan"
			org      = "myorg"
		)

		setFlag(&cmd, "-b", broker)
		setFlag(&cmd, "-o", org)
		setFlag(&cmd, "-p", plan)
		setPositionalFlags(&cmd, offering)

		err := cmd.Execute(nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(fakeActor.DisableServiceAccessCallCount()).To(Equal(1))
		actualOffering, actualBroker, actualOrg, actualPlan := fakeActor.DisableServiceAccessArgsForCall(0)
		Expect(actualOffering).To(Equal(offering))
		Expect(actualPlan).To(Equal(plan))
		Expect(actualOrg).To(Equal(org))
		Expect(actualBroker).To(Equal(broker))
	})

	It("says OK and reports warnings", func() {
		err := cmd.Execute(nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(testUI.Out).To(Say("OK"))
		Expect(testUI.Err).To(Say("a warning"))
	})

	When("some plans were skipped", func() {
		BeforeEach(func() {
			fakeActor.DisableServiceAccessReturns(v7action.SkippedPlans{"skipped_1", "skipped_2"}, nil, nil)
		})

		It("reports them", func() {
			err := cmd.Execute(nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(testUI.Out).To(Say("Did not update plan skipped_1 as it already has visibility none\\."))
			Expect(testUI.Out).To(Say("Did not update plan skipped_2 as it already has visibility none\\."))
			Expect(testUI.Out).To(Say("OK"))
		})
	})

	When("the actor return an error", func() {
		BeforeEach(func() {
			fakeActor.DisableServiceAccessReturns(nil, v7action.Warnings{"careful"}, errors.New("badness"))
		})

		It("fails with warnings", func() {
			err := cmd.Execute(nil)
			Expect(err).To(MatchError("badness"))
			Expect(testUI.Err).To(Say("careful"))
		})
	})

	When("not logged in", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(errors.New("not logged in"))
		})

		It("fails", func() {
			err := cmd.Execute(nil)
			Expect(err).To(MatchError("not logged in"))
		})
	})
})
