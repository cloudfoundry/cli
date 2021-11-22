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
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("set-health-check Command", func() {
	var (
		cmd             SetHealthCheckCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		binaryName      string
		executeErr      error
		app             string
		healthCheckType constant.HealthCheckType
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		app = "some-app"
		healthCheckType = "some-health-check-type"

		cmd = SetHealthCheckCommand{
			RequiredArgs:      flag.SetHealthCheckArgs{AppName: app, HealthCheck: flag.HealthCheckType{Type: healthCheckType}},
			HTTPEndpoint:      "some-http-endpoint",
			ProcessType:       "some-process-type",
			InvocationTimeout: flag.PositiveInteger{Value: 42},

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

	When("updating the application process health check returns an error", func() {
		var expectedErr error

		BeforeEach(func() {
			expectedErr = actionerror.ApplicationNotFoundError{Name: app}
			fakeActor.SetApplicationProcessHealthCheckTypeByNameAndSpaceReturns(resources.Application{}, v7action.Warnings{"warning-1", "warning-2"}, expectedErr)
		})

		It("returns the error and prints warnings", func() {
			Expect(executeErr).To(Equal(actionerror.ApplicationNotFoundError{Name: app}))

			Expect(testUI.Out).To(Say(`Updating health check type for app some-app process some-process-type in org some-org / space some-space as steve\.\.\.`))

			Expect(testUI.Err).To(Say("warning-1"))
			Expect(testUI.Err).To(Say("warning-2"))
		})
	})

	When("application is started", func() {
		BeforeEach(func() {
			fakeActor.SetApplicationProcessHealthCheckTypeByNameAndSpaceReturns(
				resources.Application{
					State: constant.ApplicationStarted,
				},
				v7action.Warnings{"warning-1", "warning-2"},
				nil)
		})

		It("displays a message to restart application", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(testUI.Out).To(Say(`Updating health check type for app some-app process some-process-type in org some-org / space some-space as steve\.\.\.`))
			Expect(testUI.Out).To(Say(`TIP: An app restart is required for the change to take effect\.`))

			Expect(fakeActor.SetApplicationProcessHealthCheckTypeByNameAndSpaceCallCount()).To(Equal(1))
			appName, spaceGUID, healthCheckType, httpEndpoint, processType, invocationTimeout := fakeActor.SetApplicationProcessHealthCheckTypeByNameAndSpaceArgsForCall(0)
			Expect(appName).To(Equal("some-app"))
			Expect(spaceGUID).To(Equal("some-space-guid"))
			Expect(healthCheckType).To(Equal(constant.HealthCheckType("some-health-check-type")))
			Expect(httpEndpoint).To(Equal("some-http-endpoint"))
			Expect(processType).To(Equal("some-process-type"))
			Expect(invocationTimeout).To(BeEquivalentTo(42))

			Expect(testUI.Err).To(Say("warning-1"))
			Expect(testUI.Err).To(Say("warning-2"))
		})
	})

	When("app is not started", func() {
		BeforeEach(func() {
			fakeActor.SetApplicationProcessHealthCheckTypeByNameAndSpaceReturns(
				resources.Application{
					State: constant.ApplicationStopped,
				},
				v7action.Warnings{"warning-1", "warning-2"},
				nil)
		})

		It("does not display a message to restart application", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(testUI.Out).To(Say(`Updating health check type for app some-app process some-process-type in org some-org / space some-space as steve\.\.\.`))
			Expect(testUI.Out).NotTo(Say(`TIP: An app restart is required for the change to take effect\.`))

			Expect(testUI.Err).To(Say("warning-1"))
			Expect(testUI.Err).To(Say("warning-2"))
		})
	})
})
