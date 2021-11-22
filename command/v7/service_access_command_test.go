package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
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

var _ = Describe("service-access Command", func() {
	var (
		cmd             ServiceAccessCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		binaryName      string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		cmd = ServiceAccessCommand{
			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}
	})

	When("logged in", func() {
		BeforeEach(func() {
			fakeActor.GetCurrentUserReturns(configv3.User{Name: "some-user"}, nil)
			fakeConfig.TargetReturns("some-url")
		})

		DescribeTable("message text",
			func(broker, serviceOffering, organization, expectedOutput string) {

				setFlag(&cmd, "-o", organization)
				setFlag(&cmd, "-e", serviceOffering)
				setFlag(&cmd, "-b", broker)

				executeErr := cmd.Execute(nil)

				Expect(executeErr).NotTo(HaveOccurred())
				Expect(testUI.Out).To(Say(expectedOutput))
			},
			Entry("when no flags are passed", "", "", "",
				"Getting service access as some-user\\.\\.\\."),
			Entry("when the broker flag is passed", "test-broker", "", "",
				"Getting service access for broker test-broker as some-user\\.\\.\\."),
			Entry("when the broker and service flags are passed", "test-broker", "test-service", "",
				"Getting service access for broker test-broker and service offering test-service as some-user\\.\\.\\."),
			Entry("when the broker and org flags are passed", "test-broker", "", "test-org",
				"Getting service access for broker test-broker and organization test-org as some-user\\.\\.\\."),
			Entry("when the broker, service and org flags are passed", "test-broker", "test-service", "test-org",
				"Getting service access for broker test-broker and service offering test-service and organization test-org as some-user\\.\\.\\."),
			Entry("when the service flag is passed", "", "test-service", "",
				"Getting service access for service offering test-service as some-user\\.\\.\\."),
			Entry("when the service and org flags are passed", "", "test-service", "test-org",
				"Getting service access for service offering test-service and organization test-org as some-user\\.\\.\\."),
			Entry("when the org flag is passed", "", "", "test-org",
				"Getting service access for organization test-org as some-user\\.\\.\\."),
		)

		When("there are service plans", func() {
			BeforeEach(func() {
				fakeActor.GetServiceAccessReturns(
					fakeServiceAccessResult(),
					v7action.Warnings{"warning"},
					nil,
				)
			})

			It("displays broker, service plan and access", func() {
				executeErr := cmd.Execute(nil)
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(testUI.Err).To(Say("warning"))

				const tableHeaders = `offering\s+plan\s+access\s+orgs`
				Expect(testUI.Out).To(Say(`broker:\s+broker-one`))
				Expect(testUI.Out).To(Say(tableHeaders))
				Expect(testUI.Out).To(Say(`service-one\s+plan-one\s+all`))
				Expect(testUI.Out).To(Say(`service-two\s+plan-two\s+none`))

				Expect(testUI.Out).To(Say(`broker:\s+broker-two`))
				Expect(testUI.Out).To(Say(tableHeaders))
				Expect(testUI.Out).To(Say(`service-three\s+plan-three\s+limited\s+org-1,org-2`))
				Expect(testUI.Out).To(Say(`service-four\s+plan-four\s+limited\s+org-1,org-3`))

				Expect(testUI.Out).To(Say(`broker:\s+broker-three`))
				Expect(testUI.Out).To(Say(`offering\s+plan\s+access\s+space`))
				Expect(testUI.Out).To(Say(`service-five\s+plan-five\s+limited\s+space-1 in org org-1`))
			})
		})

		When("there are no service plans", func() {
			BeforeEach(func() {
				fakeActor.GetServiceAccessReturns(
					[]v7action.ServicePlanAccess{},
					v7action.Warnings{"a warning"},
					nil,
				)
			})

			It("displays a message", func() {
				executeErr := cmd.Execute(nil)
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(testUI.Err).To(Say("a warning"))
				Expect(testUI.Out).To(Say("No service plans found."))
			})
		})

		When("resource flags are passed", func() {
			BeforeEach(func() {
				setFlag(&cmd, "-o", "test-organization")
				setFlag(&cmd, "-e", "test-service")
				setFlag(&cmd, "-b", "test-broker")
			})

			It("passes the right flags to the actor", func() {
				executeErr := cmd.Execute(nil)
				Expect(executeErr).NotTo(HaveOccurred())

				Expect(fakeActor.GetServiceAccessCallCount()).To(Equal(1))

				actualService, actualBroker, actualOrg := fakeActor.GetServiceAccessArgsForCall(0)
				Expect(actualBroker).To(Equal("test-broker"))
				Expect(actualService).To(Equal("test-service"))
				Expect(actualOrg).To(Equal("test-organization"))
			})

		})

		When("the actor errors", func() {
			BeforeEach(func() {
				fakeActor.GetServiceAccessReturns(nil, v7action.Warnings{"warning"}, errors.New("explode"))
			})

			It("displays warnings and propagates the error", func() {
				executeErr := cmd.Execute(nil)
				Expect(testUI.Err).To(Say("warning"))
				Expect(executeErr).To(MatchError("explode"))
			})
		})
	})

	When("checking target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			executeErr := cmd.Execute(nil)
			Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: binaryName}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeFalse())
			Expect(checkTargetedSpace).To(BeFalse())
		})
	})

	When("getting user fails", func() {
		BeforeEach(func() {
			fakeActor.GetCurrentUserReturns(configv3.User{}, errors.New("fake get user error"))
		})

		It("returns an error", func() {
			executeErr := cmd.Execute(nil)
			Expect(executeErr).To(MatchError("fake get user error"))

			Expect(fakeActor.GetCurrentUserCallCount()).To(Equal(1))
		})
	})
})

func fakeServiceAccessResult() []v7action.ServicePlanAccess {
	return []v7action.ServicePlanAccess{
		{
			BrokerName:          "broker-one",
			ServiceOfferingName: "service-one",
			ServicePlanName:     "plan-one",
			VisibilityType:      "public",
		},
		{
			BrokerName:          "broker-one",
			ServiceOfferingName: "service-two",
			ServicePlanName:     "plan-two",
			VisibilityType:      "admin",
		},
		{
			BrokerName:          "broker-two",
			ServiceOfferingName: "service-three",
			ServicePlanName:     "plan-three",
			VisibilityType:      "organization",
			VisibilityDetails:   []string{"org-1", "org-2"},
		},
		{
			BrokerName:          "broker-two",
			ServiceOfferingName: "service-four",
			ServicePlanName:     "plan-four",
			VisibilityType:      "organization",
			VisibilityDetails:   []string{"org-1", "org-3"},
		},
		{
			BrokerName:          "broker-three",
			ServiceOfferingName: "service-five",
			ServicePlanName:     "plan-five",
			VisibilityType:      "space",
			VisibilityDetails:   []string{"space-1 in org org-1"},
		},
	}
}
