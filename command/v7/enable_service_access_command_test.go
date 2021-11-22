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

var _ = Describe("enable-service-access command", func() {
	var (
		cmd             EnableServiceAccessCommand
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

		cmd = EnableServiceAccessCommand{
			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		setPositionalFlags(&cmd, "some-service")
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
			fakeActor.GetCurrentUserReturns(configv3.User{Name: "fake-user"}, nil)

			setPositionalFlags(&cmd, "fake-service")
			setFlag(&cmd, "-p", plan)
			setFlag(&cmd, "-o", org)
			setFlag(&cmd, "-b", broker)

			err := cmd.Execute(nil)
			Expect(err).NotTo(HaveOccurred())

			Expect(testUI.Out).To(Say(expected))
		},
		Entry("no flags", "", "", "",
			`Enabling access to all plans of service offering fake-service for all orgs as fake-user\.\.\.`),
		Entry("plan", "fake-plan", "", "",
			`Enabling access to plan fake-plan of service offering fake-service for all orgs as fake-user\.\.\.`),
		Entry("org", "", "fake-org", "",
			`Enabling access to all plans of service offering fake-service for org fake-org as fake-user\.\.\.`),
		Entry("broker", "", "", "fake-broker",
			`Enabling access to all plans of service offering fake-service from broker fake-broker for all orgs as fake-user\.\.\.`),
		Entry("plan and org", "fake-plan", "fake-org", "",
			`Enabling access to plan fake-plan of service offering fake-service for org fake-org as fake-user\.\.\.`),
		Entry("plan and broker", "fake-plan", "", "fake-broker",
			`Enabling access to plan fake-plan of service offering fake-service from broker fake-broker for all orgs as fake-user\.\.\.`),
		Entry("plan, org and broker", "fake-plan", "fake-org", "fake-broker",
			`Enabling access to plan fake-plan of service offering fake-service from broker fake-broker for org fake-org as fake-user\.\.\.`),
		Entry("broker and org", "", "fake-org", "fake-broker",
			`Enabling access to all plans of service offering fake-service from broker fake-broker for org fake-org as fake-user\.\.\.`),
	)

	It("calls the actor with the right arguments", func() {
		const (
			offeringName = "some-offering"
			planName     = "some-plan"
			orgName      = "some-org"
			brokerName   = "some-broker"
		)

		setFlag(&cmd, "-b", brokerName)
		setFlag(&cmd, "-o", orgName)
		setFlag(&cmd, "-p", planName)
		setPositionalFlags(&cmd, offeringName)

		fakeActor.EnableServiceAccessReturns(v7action.SkippedPlans{}, v7action.Warnings{"a warning"}, nil)

		err := cmd.Execute(nil)

		Expect(err).NotTo(HaveOccurred())
		Expect(testUI.Out).To(Say("OK"))
		Expect(testUI.Err).To(Say("a warning"))

		Expect(fakeActor.EnableServiceAccessCallCount()).To(Equal(1))

		actualOfferingName, actualBrokerName, actualOrgName, actualPlanName := fakeActor.EnableServiceAccessArgsForCall(0)
		Expect(actualOfferingName).To(Equal(offeringName))
		Expect(actualPlanName).To(Equal(planName))
		Expect(actualOrgName).To(Equal(orgName))
		Expect(actualBrokerName).To(Equal(brokerName))
	})

	It("reports on skipped plans", func() {
		const offeringName = "some-offering"
		setPositionalFlags(&cmd, offeringName)

		fakeActor.EnableServiceAccessReturns(
			v7action.SkippedPlans{"skipped_1", "skipped_2"},
			v7action.Warnings{"a warning"},
			nil,
		)

		err := cmd.Execute(nil)

		Expect(err).NotTo(HaveOccurred())
		Expect(testUI.Out).To(Say("Did not update plan skipped_1 as it already has visibility all\\."))
		Expect(testUI.Out).To(Say("Did not update plan skipped_2 as it already has visibility all\\."))
		Expect(testUI.Out).To(Say("OK"))
	})

	When("the actor fails", func() {
		It("prints the error", func() {
			fakeActor.EnableServiceAccessReturns(v7action.SkippedPlans{}, v7action.Warnings{"a warning"}, errors.New("access error"))

			err := cmd.Execute(nil)
			Expect(err).To(MatchError("access error"))
			Expect(testUI.Err).To(Say("a warning"))
		})
	})

	When("checking the target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(errors.New("unable to check target"))
		})

		It("returns an error", func() {
			err := cmd.Execute(nil)
			Expect(err).To(MatchError("unable to check target"))
		})
	})
})
