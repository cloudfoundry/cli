package v6_test

import (
	"errors"
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v6"
	"code.cloudfoundry.org/cli/command/v6/v6fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

func createBroker(i int) v2action.ServiceBrokerSummary {
	return v2action.ServiceBrokerSummary{
		ServiceBroker: v2action.ServiceBroker{
			Name: fmt.Sprintf("sb%d", i),
		},
		Services: []v2action.ServiceSummary{
			{
				Service: v2action.Service{Label: fmt.Sprintf("service%d-2", i)},
				Plans: []v2action.ServicePlanSummary{
					{
						ServicePlan: v2action.ServicePlan{Name: "simple"},
						VisibleTo:   []string{"org1", "org2"},
					},
					{
						ServicePlan: v2action.ServicePlan{
							Name:   "complex",
							Public: true,
						},
					},
				},
			},
			{
				Service: v2action.Service{Label: fmt.Sprintf("service%d-1", i)},
				Plans: []v2action.ServicePlanSummary{
					{
						ServicePlan: v2action.ServicePlan{Name: "simple"},
					},
					{
						ServicePlan: v2action.ServicePlan{
							Name:   "complex",
							Public: true,
						},
						VisibleTo: []string{"org3", "org4"},
					},
				},
			},
		},
	}
}

func rowMatcher(brokers []v2action.ServiceBrokerSummary, b int, s int, p int, access string) string {
	row := fmt.Sprintf(
		`\s+%s\s+%s\s+%s`,
		brokers[b].Services[s].Label,
		brokers[b].Services[s].Plans[p].Name,
		access,
	)

	if len(brokers[b].Services[s].Plans[p].VisibleTo) > 0 && access != "all" {
		row = fmt.Sprintf(`%s\s+%s`, row, strings.Join(brokers[b].Services[s].Plans[p].VisibleTo, ","))
	}

	return row
}

var _ = Describe("service-access Command", func() {
	var (
		cmd             ServiceAccessCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v6fakes.FakeServiceAccessActor
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v6fakes.FakeServiceAccessActor)

		cmd = ServiceAccessCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns("faceman")
	})

	When("a cloud controller API endpoint is set", func() {
		BeforeEach(func() {
			fakeConfig.TargetReturns("some-url")
		})

		When("checking target fails", func() {
			BeforeEach(func() {
				fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: binaryName})
			})

			It("returns an error", func() {
				executeErr = cmd.Execute(nil)
				Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: binaryName}))

				Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
				checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
				Expect(checkTargetedOrg).To(BeFalse())
				Expect(checkTargetedSpace).To(BeFalse())
			})
		})

		When("the user is logged in", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserReturns(
					configv3.User{Name: "some-user"},
					nil)
			})

			DescribeTable("flavour text",
				func(broker, service, organization, expectedOutput string) {
					cmd.Broker = broker
					cmd.Service = service
					cmd.Organization = organization

					executeErr = cmd.Execute(nil)

					Expect(executeErr).NotTo(HaveOccurred())
					Expect(testUI.Out).To(Say(expectedOutput))
				},
				Entry("when no flags are passed", "", "", "",
					"Getting service access as some-user\\.\\.\\."),
				Entry("when the broker flag is passed", "test-broker", "", "",
					"Getting service access for broker test-broker as some-user\\.\\.\\."),
				Entry("when the broker and service flags are passed", "test-broker", "test-service", "",
					"Getting service access for broker test-broker and service test-service as some-user\\.\\.\\."),
				Entry("when the broker and org flags are passed", "test-broker", "", "test-org",
					"Getting service access for broker test-broker and organization test-org as some-user\\.\\.\\."),
				Entry("when the broker, service and org flags are passed", "test-broker", "test-service", "test-org",
					"Getting service access for broker test-broker and service test-service and organization test-org as some-user\\.\\.\\."),
				Entry("when the service flag is passed", "", "test-service", "",
					"Getting service access for service test-service as some-user\\.\\.\\."),
				Entry("when the service and org flags are passed", "", "test-service", "test-org",
					"Getting service access for service test-service and organization test-org as some-user\\.\\.\\."),
				Entry("when the org flag is passed", "", "", "test-org",
					"Getting service access for organization test-org as some-user\\.\\.\\."),
			)

			When("there are no broker summaries returned", func() {
				BeforeEach(func() {
					fakeActor.GetServiceBrokerSummariesReturns([]v2action.ServiceBrokerSummary{}, nil, nil)
				})

				It("displays only the header and nothing else", func() {
					executeErr = cmd.Execute(nil)
					Expect(executeErr).NotTo(HaveOccurred())
					Eventually(testUI.Out).Should(Say("Getting service access as some-user\\.\\.\\."))
					Consistently(testUI.Out).ShouldNot(Say("[^\\s]"))
				})
			})

			When("flags are passed", func() {
				BeforeEach(func() {
					cmd.Broker = "test-broker"
					cmd.Service = "test-service"
					cmd.Organization = "test-organization"
				})

				JustBeforeEach(func() {
					executeErr = cmd.Execute(nil)
				})

				It("fetches service broker summaries from the passed flags", func() {
					Expect(executeErr).NotTo(HaveOccurred())

					Expect(fakeActor.GetServiceBrokerSummariesCallCount()).To(Equal(1))

					actualBroker, actualService, actualOrg := fakeActor.GetServiceBrokerSummariesArgsForCall(0)
					Expect(actualBroker).To(Equal("test-broker"))
					Expect(actualService).To(Equal("test-service"))
					Expect(actualOrg).To(Equal("test-organization"))
				})

				Context("but fetching summaries fails", func() {
					BeforeEach(func() {
						fakeActor.GetServiceBrokerSummariesReturns(nil, v2action.Warnings{"warning"}, errors.New("explode"))
					})

					It("displays warnings and propagates the error", func() {
						Expect(testUI.Err).To(Say("warning"))
						Expect(executeErr).To(MatchError("explode"))
					})
				})
			})

			Describe("tabular output", func() {
				var brokers []v2action.ServiceBrokerSummary
				JustBeforeEach(func() {
					executeErr = cmd.Execute(nil)
				})

				When("the summaries returned are unordered", func() {
					BeforeEach(func() {
						brokers = []v2action.ServiceBrokerSummary{createBroker(2), createBroker(1)}
						fakeActor.GetServiceBrokerSummariesReturns(brokers, v2action.Warnings{"warning"}, nil)
					})

					It("sorts brokers and services before displaying broker, service plan and access with org", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(testUI.Err).To(Say("warning"))

						// The command sorts the slices in places, so the order here is not the same as the order they were
						// generated in and passed to GetServiceBrokerSummariesReturns.
						tableHeaders := `service\s+plan\s+access\s+orgs`
						Expect(testUI.Out).To(Say(`broker:\s+%s`, brokers[0].Name))
						Expect(testUI.Out).To(Say(tableHeaders))
						Expect(testUI.Out).To(Say(rowMatcher(brokers, 0, 0, 0, "all")))
						Expect(testUI.Out).To(Say(rowMatcher(brokers, 0, 0, 1, "none")))
						Expect(testUI.Out).To(Say(rowMatcher(brokers, 0, 1, 0, "all")))
						Expect(testUI.Out).To(Say(rowMatcher(brokers, 0, 1, 1, "limited")))
						Expect(testUI.Out).To(Say(`broker:\s+%s`, brokers[1].Name))
						Expect(testUI.Out).To(Say(tableHeaders))
						Expect(testUI.Out).To(Say(rowMatcher(brokers, 1, 0, 0, "all")))
						Expect(testUI.Out).To(Say(rowMatcher(brokers, 1, 0, 1, "none")))
						Expect(testUI.Out).To(Say(rowMatcher(brokers, 1, 1, 0, "all")))
						Expect(testUI.Out).To(Say(rowMatcher(brokers, 1, 1, 1, "limited")))
					})
				})
			})
		})
	})
})
