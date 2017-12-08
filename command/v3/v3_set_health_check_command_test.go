package v3_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v3"
	"code.cloudfoundry.org/cli/command/v3/v3fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("v3-set-health-check Command", func() {
	var (
		cmd             v3.V3SetHealthCheckCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v3fakes.FakeV3SetHealthCheckActor
		binaryName      string
		executeErr      error
		app             string
		healthCheckType string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v3fakes.FakeV3SetHealthCheckActor)

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		app = "some-app"
		healthCheckType = "some-health-check-type"

		cmd = v3.V3SetHealthCheckCommand{
			RequiredArgs: flag.SetHealthCheckArgs{AppName: app, HealthCheck: flag.HealthCheckType{Type: healthCheckType}},
			HTTPEndpoint: "some-http-endpoint",
			ProcessType:  "some-process-type",

			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
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
		fakeActor.CloudControllerAPIVersionReturns(ccversion.MinVersionV3)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	Context("when the API version is below the minimum", func() {
		BeforeEach(func() {
			fakeActor.CloudControllerAPIVersionReturns("0.0.0")
		})

		It("returns a MinimumAPIVersionNotMetError", func() {
			Expect(executeErr).To(MatchError(translatableerror.MinimumAPIVersionNotMetError{
				CurrentVersion: "0.0.0",
				MinimumVersion: ccversion.MinVersionV3,
			}))
		})

		It("displays the experimental warning", func() {
			Expect(testUI.Out).To(Say("This command is in EXPERIMENTAL stage and may change without notice"))
		})
	})

	Context("when checking target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NoOrganizationTargetedError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NoOrganizationTargetedError{BinaryName: binaryName}))

			Expect(testUI.Out).To(Say("This command is in EXPERIMENTAL stage and may change without notice"))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeTrue())
			Expect(checkTargetedSpace).To(BeTrue())
		})
	})

	Context("when the user is not logged in", func() {
		var expectedErr error

		BeforeEach(func() {
			expectedErr = errors.New("some current user error")
			fakeConfig.CurrentUserReturns(configv3.User{}, expectedErr)
		})

		It("return an error", func() {
			Expect(executeErr).To(Equal(expectedErr))
		})
	})

	Context("when updating the application process health check returns an error", func() {
		var expectedErr error

		BeforeEach(func() {
			expectedErr = actionerror.ApplicationNotFoundError{Name: app}
			fakeActor.SetApplicationProcessHealthCheckTypeByNameAndSpaceReturns(v3action.Application{}, v3action.Warnings{"warning-1", "warning-2"}, expectedErr)
		})

		It("returns the error and prints warnings", func() {
			Expect(executeErr).To(Equal(actionerror.ApplicationNotFoundError{Name: app}))

			Expect(testUI.Out).To(Say("This command is in EXPERIMENTAL stage and may change without notice"))
			Expect(testUI.Out).To(Say("Updating health check type for app some-app process some-process-type in org some-org / space some-space as steve\\.\\.\\."))

			Expect(testUI.Err).To(Say("warning-1"))
			Expect(testUI.Err).To(Say("warning-2"))
		})
	})

	Context("when application is started", func() {
		BeforeEach(func() {
			fakeActor.SetApplicationProcessHealthCheckTypeByNameAndSpaceReturns(
				v3action.Application{
					State: constant.ApplicationStarted,
				},
				v3action.Warnings{"warning-1", "warning-2"},
				nil)
		})

		It("displays a message to restart application", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(testUI.Out).To(Say("This command is in EXPERIMENTAL stage and may change without notice"))
			Expect(testUI.Out).To(Say("Updating health check type for app some-app process some-process-type in org some-org / space some-space as steve\\.\\.\\."))
			Expect(testUI.Out).To(Say("TIP: An app restart is required for the change to take effect\\."))

			Expect(fakeActor.SetApplicationProcessHealthCheckTypeByNameAndSpaceCallCount()).To(Equal(1))
			appName, spaceGUID, healthCheckType, httpEndpoint, processType := fakeActor.SetApplicationProcessHealthCheckTypeByNameAndSpaceArgsForCall(0)
			Expect(appName).To(Equal("some-app"))
			Expect(spaceGUID).To(Equal("some-space-guid"))
			Expect(healthCheckType).To(Equal("some-health-check-type"))
			Expect(httpEndpoint).To(Equal("some-http-endpoint"))
			Expect(processType).To(Equal("some-process-type"))

			Expect(testUI.Err).To(Say("warning-1"))
			Expect(testUI.Err).To(Say("warning-2"))
		})
	})

	Context("when app is not started", func() {
		BeforeEach(func() {
			fakeActor.SetApplicationProcessHealthCheckTypeByNameAndSpaceReturns(
				v3action.Application{
					State: constant.ApplicationStopped,
				},
				v3action.Warnings{"warning-1", "warning-2"},
				nil)
		})

		It("does not display a message to restart application", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(testUI.Out).To(Say("This command is in EXPERIMENTAL stage and may change without notice"))
			Expect(testUI.Out).To(Say("Updating health check type for app some-app process some-process-type in org some-org / space some-space as steve\\.\\.\\."))
			Expect(testUI.Out).NotTo(Say("TIP: An app restart is required for the change to take effect\\."))

			Expect(testUI.Err).To(Say("warning-1"))
			Expect(testUI.Err).To(Say("warning-2"))
		})
	})
})
