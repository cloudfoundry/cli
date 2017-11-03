package v2_test

import (
	"errors"

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

var _ = Describe("get-health-check Command", func() {
	var (
		cmd             GetHealthCheckCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v2fakes.FakeGetHealthCheckActor
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v2fakes.FakeGetHealthCheckActor)

		cmd = GetHealthCheckCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		fakeConfig.TargetedOrganizationReturns(configv3.Organization{
			Name: "some-org",
		})
		fakeConfig.TargetedSpaceReturns(configv3.Space{
			GUID: "some-space-guid",
			Name: "some-space",
		})
		fakeConfig.CurrentUserReturns(configv3.User{Name: "some-user"}, nil)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	Context("when checking the target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(
				actionerror.NotLoggedInError{BinaryName: binaryName})
		})

		It("returns a wrapped error", func() {
			Expect(executeErr).To(MatchError(
				actionerror.NotLoggedInError{BinaryName: binaryName}))
		})
	})

	Context("when getting the user returns an error", func() {
		var expectedErr error

		BeforeEach(func() {
			expectedErr = errors.New("current user error")
			fakeConfig.CurrentUserReturns(configv3.User{}, expectedErr)
		})

		It("returns a wrapped error", func() {
			Expect(executeErr).To(MatchError(expectedErr))
		})
	})

	Context("when getting the application returns an error", func() {
		var expectedErr error

		BeforeEach(func() {
			cmd.RequiredArgs.AppName = "some-app"

			expectedErr = errors.New("get health check error")
			fakeActor.GetApplicationByNameAndSpaceReturns(
				v2action.Application{}, v2action.Warnings{"warning-1"}, expectedErr)
		})

		It("displays warnings and returns the error", func() {
			Expect(testUI.Err).To(Say("warning-1"))
			Expect(executeErr).To(MatchError(expectedErr))
		})
	})

	Context("when getting the application is successful", func() {
		Context("when the health check type is not http", func() {
			BeforeEach(func() {
				cmd.RequiredArgs.AppName = "some-app"

				fakeActor.GetApplicationByNameAndSpaceReturns(
					v2action.Application{
						HealthCheckType:         "some-health-check-type",
						HealthCheckHTTPEndpoint: "/some-endpoint",
					}, v2action.Warnings{"warning-1"}, nil)
			})

			It("show a blank endpoint and displays warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Out).To(Say("Getting health check type for app some-app in org some-org / space some-space as some-user..."))
				Expect(testUI.Out).To(Say("\n\n"))
				Expect(testUI.Out).To(Say("health check type:          some-health-check-type"))
				Expect(testUI.Out).To(Say("endpoint \\(for http type\\):   \n"))

				Expect(testUI.Err).To(Say("warning-1"))

				Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
				targetedOrganizationRequired, targetedSpaceRequired := fakeSharedActor.CheckTargetArgsForCall(0)
				Expect(targetedOrganizationRequired).To(Equal(true))
				Expect(targetedSpaceRequired).To(Equal(true))

				Expect(fakeConfig.CurrentUserCallCount()).To(Equal(1))

				Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
				name, spaceGUID := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
				Expect(name).To(Equal("some-app"))
				Expect(spaceGUID).To(Equal("some-space-guid"))
			})
		})

		Context("when the health check type is http", func() {
			BeforeEach(func() {
				cmd.RequiredArgs.AppName = "some-app"

				fakeActor.GetApplicationByNameAndSpaceReturns(
					v2action.Application{
						HealthCheckType:         "http",
						HealthCheckHTTPEndpoint: "/some-endpoint",
					}, v2action.Warnings{"warning-1"}, nil)
			})

			It("shows the endpoint", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Out).To(Say("Getting health check type for app some-app in org some-org / space some-space as some-user..."))
				Expect(testUI.Out).To(Say("\n\n"))
				Expect(testUI.Out).To(Say("health check type:          http"))
				Expect(testUI.Out).To(Say("endpoint \\(for http type\\):   /some-endpoint"))
			})
		})
	})
})
