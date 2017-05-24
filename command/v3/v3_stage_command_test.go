package v3_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/v3"
	"code.cloudfoundry.org/cli/command/v3/v3fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("v3-stage Command", func() {
	var (
		cmd             v3.V3StageCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v3fakes.FakeV3StageActor
		binaryName      string
		executeErr      error
		app             string
		packageGUID     string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v3fakes.FakeV3StageActor)

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		app = "some-app"
		packageGUID = "some-package-guid"

		cmd = v3.V3StageCommand{
			AppName:     app,
			PackageGUID: packageGUID,

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
			Expect(executeErr).To(MatchError(command.NotLoggedInError{BinaryName: binaryName}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			_, checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeTrue())
			Expect(checkTargetedSpace).To(BeTrue())
		})
	})

	Context("when the user is logged in", func() {
		BeforeEach(func() {
			fakeConfig.HasTargetedOrganizationReturns(true)
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				GUID: "some-org-guid",
				Name: "some-org",
			})
			fakeConfig.HasTargetedSpaceReturns(true)
			fakeConfig.TargetedSpaceReturns(configv3.Space{
				GUID: "some-space-guid",
				Name: "some-space",
			})
			fakeConfig.CurrentUserReturns(configv3.User{Name: "steve"}, nil)

			fakeActor.StagePackageReturns(v3action.Build{Droplet: ccv3.Droplet{GUID: "some-droplet-guid"}}, v3action.Warnings{"some-warning", "some-other-warning"}, nil)
		})

		It("outputs the droplet GUID", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(testUI.Out).To(Say("Staging package for %s in org some-org / space some-space as steve...", app))
			Expect(testUI.Out).To(Say("droplet: some-droplet-guid"))
			Expect(testUI.Out).To(Say("OK"))

			Expect(testUI.Err).To(Say("some-warning"))
			Expect(testUI.Err).To(Say("some-other-warning"))
		})

		It("stages the package", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(fakeActor.StagePackageCallCount()).To(Equal(1))
			Expect(fakeActor.StagePackageArgsForCall(0)).To(Equal(packageGUID))
		})

		Context("when the staging returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("any gibberish")
				fakeActor.StagePackageReturns(v3action.Build{}, v3action.Warnings{"some-warning", "some-other-warning"}, expectedErr)
			})

			It("returns the error and displays warnings", func() {
				Expect(executeErr).To(Equal(expectedErr))

				Expect(testUI.Err).To(Say("some-warning"))
				Expect(testUI.Err).To(Say("some-other-warning"))
			})
		})
	})
})
