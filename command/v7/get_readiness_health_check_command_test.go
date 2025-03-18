package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/v8/actor/actionerror"
	"code.cloudfoundry.org/cli/v8/actor/v7action"
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/v8/command/commandfakes"
	"code.cloudfoundry.org/cli/v8/command/flag"
	. "code.cloudfoundry.org/cli/v8/command/v7"
	"code.cloudfoundry.org/cli/v8/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/v8/util/configv3"
	"code.cloudfoundry.org/cli/v8/util/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("get-readiness-health-check Command", func() {
	var (
		cmd             GetReadinessHealthCheckCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		binaryName      string
		executeErr      error
		app             string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		app = "some-app"

		cmd = GetReadinessHealthCheckCommand{
			RequiredArgs: flag.AppName{AppName: app},

			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		fakeConfig.TargetedOrganizationReturns(configv3.Organization{
			Name: "some-org",
			GUID: "some-org-guid",
		})
		fakeConfig.TargetedSpaceReturns(configv3.Space{
			Name: "some-space",
			GUID: "some-space-guid",
		})

		fakeActor.GetCurrentUserReturns(configv3.User{Name: "steve"}, nil)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("checking target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NoOrganizationTargetedError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NoOrganizationTargetedError{BinaryName: binaryName}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeTrue())
			Expect(checkTargetedSpace).To(BeTrue())
		})
	})

	When("the user is not logged in", func() {
		var expectedErr error

		BeforeEach(func() {
			expectedErr = errors.New("some current user error")
			fakeActor.GetCurrentUserReturns(configv3.User{}, expectedErr)
		})

		It("return an error", func() {
			Expect(executeErr).To(Equal(expectedErr))
		})
	})

	When("getting the application process readiness health checks returns an error", func() {
		var expectedErr error

		BeforeEach(func() {
			expectedErr = actionerror.ApplicationNotFoundError{Name: app}
			fakeActor.GetApplicationProcessReadinessHealthChecksByNameAndSpaceReturns(nil, v7action.Warnings{"warning-1", "warning-2"}, expectedErr)
		})

		It("returns the error and prints warnings", func() {
			Expect(executeErr).To(Equal(actionerror.ApplicationNotFoundError{Name: app}))

			Expect(testUI.Out).To(Say("Getting readiness health check type for app some-app in org some-org / space some-space as steve..."))

			Expect(testUI.Err).To(Say("warning-1"))
			Expect(testUI.Err).To(Say("warning-2"))
		})
	})

	When("app has no processes", func() {
		BeforeEach(func() {
			fakeActor.GetApplicationProcessReadinessHealthChecksByNameAndSpaceReturns(
				[]v7action.ProcessReadinessHealthCheck{},
				v7action.Warnings{"warning-1", "warning-2"},
				nil)
		})

		It("displays a message that there are no processes", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(testUI.Out).To(Say("Getting readiness health check type for app some-app in org some-org / space some-space as steve..."))
			Expect(testUI.Out).To(Say("App has no processes"))

			Expect(fakeActor.GetApplicationProcessReadinessHealthChecksByNameAndSpaceCallCount()).To(Equal(1))
			appName, spaceGUID := fakeActor.GetApplicationProcessReadinessHealthChecksByNameAndSpaceArgsForCall(0)
			Expect(appName).To(Equal("some-app"))
			Expect(spaceGUID).To(Equal("some-space-guid"))
		})
	})

	When("app has processes", func() {
		BeforeEach(func() {
			appProcessReadinessHealthChecks := []v7action.ProcessReadinessHealthCheck{
				{ProcessType: constant.ProcessTypeWeb, HealthCheckType: constant.HTTP, Endpoint: "/foo", InvocationTimeout: 10, Interval: 2},
				{ProcessType: "queue", HealthCheckType: constant.Port, Endpoint: "", InvocationTimeout: 0},
				{ProcessType: "timer", HealthCheckType: constant.Process, Endpoint: "", InvocationTimeout: 5},
			}
			fakeActor.GetApplicationProcessReadinessHealthChecksByNameAndSpaceReturns(appProcessReadinessHealthChecks, v7action.Warnings{"warning-1", "warning-2"}, nil)
		})

		It("prints the readiness health check type of each process and warnings", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(testUI.Out).To(Say("Getting readiness health check type for app some-app in org some-org / space some-space as steve..."))
			Expect(testUI.Out).To(Say(`process\s+type\s+endpoint\s+\(for http\)\s+invocation timeout\s+interval\n`))
			Expect(testUI.Out).To(Say(`web\s+http\s+/foo\s+10\s+2\n`))
			Expect(testUI.Out).To(Say(`queue\s+port\s+\n`))
			Expect(testUI.Out).To(Say(`timer\s+process\s+5\s+\n`))

			Expect(fakeActor.GetApplicationProcessReadinessHealthChecksByNameAndSpaceCallCount()).To(Equal(1))
			appName, spaceGUID := fakeActor.GetApplicationProcessReadinessHealthChecksByNameAndSpaceArgsForCall(0)
			Expect(appName).To(Equal("some-app"))
			Expect(spaceGUID).To(Equal("some-space-guid"))
		})
	})
})
