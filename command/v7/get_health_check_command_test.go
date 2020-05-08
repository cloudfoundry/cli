package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("get-health-check Command", func() {
	var (
		cmd             GetHealthCheckCommand
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

		cmd = GetHealthCheckCommand{
			RequiredArgs: flag.AppName{AppName: app},

			BaseCommand: command.BaseCommand{
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

		fakeConfig.CurrentUserReturns(configv3.User{Name: "steve"}, nil)
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
			fakeConfig.CurrentUserReturns(configv3.User{}, expectedErr)
		})

		It("return an error", func() {
			Expect(executeErr).To(Equal(expectedErr))
		})
	})

	When("getting the application process health checks returns an error", func() {
		var expectedErr error

		BeforeEach(func() {
			expectedErr = actionerror.ApplicationNotFoundError{Name: app}
			fakeActor.GetApplicationProcessHealthChecksByNameAndSpaceReturns(nil, v7action.Warnings{"warning-1", "warning-2"}, expectedErr)
		})

		It("returns the error and prints warnings", func() {
			Expect(executeErr).To(Equal(actionerror.ApplicationNotFoundError{Name: app}))

			Expect(testUI.Out).To(Say("Getting health check type for app some-app in org some-org / space some-space as steve..."))

			Expect(testUI.Err).To(Say("warning-1"))
			Expect(testUI.Err).To(Say("warning-2"))
		})
	})

	When("app has no processes", func() {
		BeforeEach(func() {
			fakeActor.GetApplicationProcessHealthChecksByNameAndSpaceReturns(
				[]v7action.ProcessHealthCheck{},
				v7action.Warnings{"warning-1", "warning-2"},
				nil)
		})

		It("displays a message that there are no processes", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(testUI.Out).To(Say("Getting health check type for app some-app in org some-org / space some-space as steve..."))
			Expect(testUI.Out).To(Say("App has no processes"))

			Expect(fakeActor.GetApplicationProcessHealthChecksByNameAndSpaceCallCount()).To(Equal(1))
			appName, spaceGUID := fakeActor.GetApplicationProcessHealthChecksByNameAndSpaceArgsForCall(0)
			Expect(appName).To(Equal("some-app"))
			Expect(spaceGUID).To(Equal("some-space-guid"))
		})
	})

	When("app has processes", func() {
		BeforeEach(func() {
			appProcessHealthChecks := []v7action.ProcessHealthCheck{
				{ProcessType: constant.ProcessTypeWeb, HealthCheckType: constant.HTTP, Endpoint: "/foo", InvocationTimeout: 10},
				{ProcessType: "queue", HealthCheckType: constant.Port, Endpoint: "", InvocationTimeout: 0},
				{ProcessType: "timer", HealthCheckType: constant.Process, Endpoint: "", InvocationTimeout: 5},
			}
			fakeActor.GetApplicationProcessHealthChecksByNameAndSpaceReturns(appProcessHealthChecks, v7action.Warnings{"warning-1", "warning-2"}, nil)
		})

		It("prints the health check type of each process and warnings", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(testUI.Out).To(Say("Getting health check type for app some-app in org some-org / space some-space as steve..."))
			Expect(testUI.Out).To(Say(`process\s+health check\s+endpoint\s+\(for http\)\s+invocation timeout\n`))
			Expect(testUI.Out).To(Say(`web\s+http\s+/foo\s+10\n`))
			Expect(testUI.Out).To(Say(`queue\s+port\s+1\n`))
			Expect(testUI.Out).To(Say(`timer\s+process\s+5\n`))

			Expect(fakeActor.GetApplicationProcessHealthChecksByNameAndSpaceCallCount()).To(Equal(1))
			appName, spaceGUID := fakeActor.GetApplicationProcessHealthChecksByNameAndSpaceArgsForCall(0)
			Expect(appName).To(Equal("some-app"))
			Expect(spaceGUID).To(Equal("some-space-guid"))
		})
	})
})
