package v3_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v3action"
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

var _ = Describe("v3-stop Command", func() {
	var (
		cmd             v3.V3StopCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v3fakes.FakeV3StopActor
		binaryName      string
		executeErr      error
		app             string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v3fakes.FakeV3StopActor)

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		app = "some-app"

		cmd = v3.V3StopCommand{
			RequiredArgs: flag.AppName{AppName: app},

			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	Context("when checking target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(sharedaction.NoOrganizationTargetedError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(translatableerror.NoOrganizationTargetedError{BinaryName: binaryName}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			_, checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
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

	Context("when the actor does not return an error", func() {
		BeforeEach(func() {
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				Name: "some-org",
			})
			fakeConfig.TargetedSpaceReturns(configv3.Space{
				Name: "some-space",
				GUID: "some-space-guid",
			})
			fakeConfig.CurrentUserReturns(configv3.User{Name: "steve"}, nil)
			fakeActor.GetApplicationByNameAndSpaceReturns(v3action.Application{GUID: "some-app-guid", State: "STARTED"}, v3action.Warnings{"get-warning-1", "get-warning-2"}, nil)
			fakeActor.StopApplicationReturns(v3action.Warnings{"stop-warning-1", "stop-warning-2"}, nil)
		})

		It("says that the app was stopped and outputs warnings", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(testUI.Out).To(Say("Stopping app some-app in org some-org / space some-space as steve\\.\\.\\."))

			Expect(testUI.Err).To(Say("get-warning-1"))
			Expect(testUI.Err).To(Say("get-warning-2"))
			Expect(testUI.Err).To(Say("stop-warning-1"))
			Expect(testUI.Err).To(Say("stop-warning-2"))
			Expect(testUI.Out).To(Say("OK"))

			Expect(fakeActor.StopApplicationCallCount()).To(Equal(1))
			appGUID, spaceGUID := fakeActor.StopApplicationArgsForCall(0)
			Expect(appGUID).To(Equal("some-app-guid"))
			Expect(spaceGUID).To(Equal("some-space-guid"))
		})
	})

	Context("when the get app call returns a ApplicationNotFoundError", func() {
		var expectedErr error

		BeforeEach(func() {
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				Name: "some-org",
			})
			fakeConfig.TargetedSpaceReturns(configv3.Space{
				Name: "some-space",
			})
			fakeConfig.CurrentUserReturns(configv3.User{Name: "steve"}, nil)
			expectedErr = v3action.ApplicationNotFoundError{Name: app}
			fakeActor.GetApplicationByNameAndSpaceReturns(v3action.Application{State: "STARTED"}, v3action.Warnings{"get-warning-1", "get-warning-2"}, expectedErr)
		})

		It("says that the app failed to stop", func() {
			Expect(executeErr).To(Equal(translatableerror.ApplicationNotFoundError{Name: app}))
			Expect(testUI.Out).ToNot(Say("Stopping"))

			Expect(testUI.Err).To(Say("get-warning-1"))
			Expect(testUI.Err).To(Say("get-warning-2"))
		})
	})

	Context("when the stop app call returns a ApplicationNotFoundError (someone else deleted app after we fetched summary)", func() {
		var expectedErr error

		BeforeEach(func() {
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				Name: "some-org",
			})
			fakeConfig.TargetedSpaceReturns(configv3.Space{
				Name: "some-space",
			})
			fakeConfig.CurrentUserReturns(configv3.User{Name: "steve"}, nil)
			fakeActor.GetApplicationByNameAndSpaceReturns(v3action.Application{State: "STARTED"}, v3action.Warnings{"get-warning-1", "get-warning-2"}, nil)
			expectedErr = v3action.ApplicationNotFoundError{Name: app}
			fakeActor.StopApplicationReturns(v3action.Warnings{"stop-warning-1", "stop-warning-2"}, expectedErr)
		})

		It("says that the app failed to stop", func() {
			Expect(executeErr).To(Equal(translatableerror.ApplicationNotFoundError{Name: app}))
			Expect(testUI.Out).To(Say("Stopping app some-app in org some-org / space some-space as steve\\.\\.\\."))

			Expect(testUI.Err).To(Say("get-warning-1"))
			Expect(testUI.Err).To(Say("get-warning-2"))
			Expect(testUI.Err).To(Say("stop-warning-1"))
			Expect(testUI.Err).To(Say("stop-warning-2"))
		})
	})

	Context("when the app is already stopped", func() {
		BeforeEach(func() {
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				Name: "some-org",
			})
			fakeConfig.TargetedSpaceReturns(configv3.Space{
				Name: "some-space",
			})
			fakeConfig.CurrentUserReturns(configv3.User{Name: "steve"}, nil)
			fakeActor.GetApplicationByNameAndSpaceReturns(v3action.Application{State: "STOPPED"}, v3action.Warnings{"get-warning-1", "get-warning-2"}, nil)
		})

		It("says that the app failed to stop", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(testUI.Out).ToNot(Say("Stopping"))
			Expect(testUI.Out).To(Say("OK"))

			Expect(testUI.Err).To(Say("get-warning-1"))
			Expect(testUI.Err).To(Say("get-warning-2"))
			Expect(testUI.Err).To(Say("App some-app is already stopped"))

			Expect(fakeActor.StopApplicationCallCount()).To(BeZero(), "Expected StopApplication to not be called")
		})
	})

	Context("when the get application returns an unknown error", func() {
		var expectedErr error

		BeforeEach(func() {
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				Name: "some-org",
			})
			fakeConfig.TargetedSpaceReturns(configv3.Space{
				Name: "some-space",
			})
			fakeConfig.CurrentUserReturns(configv3.User{Name: "steve"}, nil)
			expectedErr = errors.New("some-error")
			fakeActor.GetApplicationByNameAndSpaceReturns(v3action.Application{State: "STARTED"}, v3action.Warnings{"get-warning-1", "get-warning-2"}, expectedErr)
		})

		It("says that the app failed to stop", func() {
			Expect(executeErr).To(Equal(expectedErr))
			Expect(testUI.Out).ToNot(Say("Stopping"))

			Expect(testUI.Err).To(Say("get-warning-1"))
			Expect(testUI.Err).To(Say("get-warning-2"))

			Expect(fakeActor.StopApplicationCallCount()).To(BeZero(), "Expected StopApplication to not be called")
		})
	})

	Context("when the stop application returns an unknown error", func() {
		var expectedErr error

		BeforeEach(func() {
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				Name: "some-org",
			})
			fakeConfig.TargetedSpaceReturns(configv3.Space{
				Name: "some-space",
			})
			fakeConfig.CurrentUserReturns(configv3.User{Name: "steve"}, nil)
			fakeActor.GetApplicationByNameAndSpaceReturns(v3action.Application{State: "STARTED"}, v3action.Warnings{"get-warning-1", "get-warning-2"}, nil)
			expectedErr = errors.New("some-error")
			fakeActor.StopApplicationReturns(v3action.Warnings{"stop-warning-1", "stop-warning-2"}, expectedErr)
		})

		It("says that the app failed to stop", func() {
			Expect(executeErr).To(Equal(expectedErr))
			Expect(testUI.Out).To(Say("Stopping app some-app in org some-org / space some-space as steve\\.\\.\\."))

			Expect(testUI.Err).To(Say("get-warning-1"))
			Expect(testUI.Err).To(Say("get-warning-2"))
			Expect(testUI.Err).To(Say("stop-warning-1"))
			Expect(testUI.Err).To(Say("stop-warning-2"))
		})
	})
})
