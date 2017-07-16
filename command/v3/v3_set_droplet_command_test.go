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

var _ = Describe("v3-set-droplet Command", func() {
	var (
		cmd             v3.V3SetDropletCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v3fakes.FakeV3SetDropletActor
		binaryName      string
		executeErr      error
		app             string
		dropletGUID     string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v3fakes.FakeV3SetDropletActor)

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		app = "some-app"
		dropletGUID = "some-droplet-guid"

		cmd = v3.V3SetDropletCommand{
			RequiredArgs: flag.AppName{AppName: app},
			DropletGUID:  dropletGUID,

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
			fakeSharedActor.CheckTargetReturns(sharedaction.NotLoggedInError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(translatableerror.NotLoggedInError{BinaryName: binaryName}))

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

	Context("when the droplet has been set to the app", func() {
		BeforeEach(func() {
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				Name: "some-org",
			})
			fakeConfig.TargetedSpaceReturns(configv3.Space{
				Name: "some-space",
				GUID: "some-space-guid",
			})
			fakeConfig.CurrentUserReturns(configv3.User{Name: "steve"}, nil)
			fakeActor.SetApplicationDropletReturns(v3action.Warnings{"warning-1", "warning-2"}, nil)
		})

		It("displays that the droplet was assigned", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(testUI.Out).To(Say("Setting app some-app to droplet some-droplet-guid in org some-org / space some-space as steve..."))

			Expect(testUI.Err).To(Say("warning-1"))
			Expect(testUI.Err).To(Say("warning-2"))
			Expect(testUI.Out).To(Say("OK"))

			Expect(fakeActor.SetApplicationDropletCallCount()).To(Equal(1))
			appName, spaceGUID, dropletGUID := fakeActor.SetApplicationDropletArgsForCall(0)
			Expect(appName).To(Equal("some-app"))
			Expect(spaceGUID).To(Equal("some-space-guid"))
			Expect(dropletGUID).To(Equal("some-droplet-guid"))
		})
	})

	Context("when the actor returns an error", func() {
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
			fakeActor.SetApplicationDropletReturns(v3action.Warnings{"warning-1", "warning-2"}, expectedErr)
		})

		It("displays that the droplet was assigned", func() {
			Expect(executeErr).To(Equal(translatableerror.ApplicationNotFoundError{Name: app}))

			Expect(testUI.Out).To(Say("Setting app some-app to droplet some-droplet-guid in org some-org / space some-space as steve..."))

			Expect(testUI.Err).To(Say("warning-1"))
			Expect(testUI.Err).To(Say("warning-2"))
		})
	})
})
