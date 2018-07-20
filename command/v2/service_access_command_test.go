package v2_test

import (
	"errors"
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v2"
	"code.cloudfoundry.org/cli/command/v2/v2fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

func generateBrokers(numberOfBrokers int) []v2action.ServiceBrokerSummary {
	var brokers []v2action.ServiceBrokerSummary
	for i := 0; i < numberOfBrokers; i++ {
		brokers = append(brokers, v2action.ServiceBrokerSummary{
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
		})
	}

	return brokers
}

func rowMatcher(brokers []v2action.ServiceBrokerSummary, b int, s int, p int, access string) string {
	row := fmt.Sprintf(
		"\\s+%s\\s+%s\\s+%s",
		brokers[b].Services[s].Label,
		brokers[b].Services[s].Plans[p].Name,
		access,
	)

	if len(brokers[b].Services[s].Plans[p].VisibleTo) > 0 && access != "all" {
		row = fmt.Sprintf("%s\\s+%s", row, strings.Join(brokers[b].Services[s].Plans[p].VisibleTo, ","))
	}

	return row
}

var _ = Describe("service-access Command", func() {
	var (
		cmd             ServiceAccessCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v2fakes.FakeServiceAccessActor
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v2fakes.FakeServiceAccessActor)

		cmd = ServiceAccessCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns("faceman")

		fakeConfig.ExperimentalReturns(true)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	Context("when a cloud controller API endpoint is set", func() {
		BeforeEach(func() {
			fakeConfig.TargetReturns("some-url")
		})

		Context("when checking target fails", func() {
			BeforeEach(func() {
				fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: binaryName})
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: binaryName}))

				Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
				checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
				Expect(checkTargetedOrg).To(BeFalse())
				Expect(checkTargetedSpace).To(BeFalse())
			})
		})

		Context("when the user is logged in", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserReturns(
					configv3.User{Name: "some-user"},
					nil)
			})

			It("displays flavor text", func() {
				Expect(testUI.Out).To(Say("Getting service access as some-user..."))
				Expect(fakeConfig.CurrentUserCallCount()).To(Equal(1))
			})

			Context("when flags are passed", func() {
				BeforeEach(func() {
					cmd.Broker = "test-broker"
					cmd.Service = "test-service"
					cmd.Organization = "test-organization"
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

			Describe("table", func() {
				var brokers []v2action.ServiceBrokerSummary

				BeforeEach(func() {
					brokers = generateBrokers(2)
					fakeActor.GetServiceBrokerSummariesReturns(generateBrokers(2), v2action.Warnings{"warning"}, nil)
				})

				It("displays each service broker, service, plan and access with org in the correct position", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					tableHeaders := "service\\s+plan\\s+access\\s+orgs"
					Expect(testUI.Out).To(Say("broker:\\s+%s", brokers[0].Name))
					Expect(testUI.Out).To(Say(tableHeaders))
					Expect(testUI.Out).To(Say(rowMatcher(brokers, 0, 1, 1, "all")))
					Expect(testUI.Out).To(Say(rowMatcher(brokers, 0, 1, 0, "none")))
					Expect(testUI.Out).To(Say(rowMatcher(brokers, 0, 0, 1, "all")))
					Expect(testUI.Out).To(Say(rowMatcher(brokers, 0, 0, 0, "limited")))
					Expect(testUI.Out).To(Say("broker:\\s+%s", brokers[1].Name))
					Expect(testUI.Out).To(Say(tableHeaders))
					Expect(testUI.Out).To(Say(rowMatcher(brokers, 1, 1, 1, "all")))
					Expect(testUI.Out).To(Say(rowMatcher(brokers, 1, 1, 0, "none")))
					Expect(testUI.Out).To(Say(rowMatcher(brokers, 1, 0, 1, "all")))
					Expect(testUI.Out).To(Say(rowMatcher(brokers, 1, 0, 0, "limited")))
				})
			})
		})
	})
})
