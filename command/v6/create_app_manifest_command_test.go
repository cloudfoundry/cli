package v6_test

import (
	"errors"
	"path/filepath"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v2v3action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	. "code.cloudfoundry.org/cli/command/v6"
	"code.cloudfoundry.org/cli/command/v6/v6fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/manifest"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("create-app-manifest Command", func() {
	var (
		cmd             CreateAppManifestCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v6fakes.FakeCreateAppManifestActor
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v6fakes.FakeCreateAppManifestActor)

		cmd = CreateAppManifestCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

		cmd.RequiredArgs.AppName = "some-app"
		cmd.FilePath = flag.Path("some-file-path")

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("checking target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: binaryName})
		})

		It("returns an error if the check fails", func() {
			Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: "faceman"}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeTrue())
			Expect(checkTargetedSpace).To(BeTrue())
		})
	})

	When("the user is logged in, and org and space are targeted", func() {
		BeforeEach(func() {
			fakeConfig.HasTargetedOrganizationReturns(true)
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "some-org"})
			fakeConfig.HasTargetedSpaceReturns(true)
			fakeConfig.TargetedSpaceReturns(configv3.Space{
				GUID: "some-space-guid",
				Name: "some-space"})
			fakeConfig.CurrentUserReturns(
				configv3.User{Name: "some-user"},
				nil)
		})

		When("creating the manifest errors", func() {
			BeforeEach(func() {
				fakeActor.CreateApplicationManifestByNameAndSpaceReturns(manifest.Application{}, v2v3action.Warnings{"some-warning"}, errors.New("some-error"))
			})

			It("returns the error, prints warnings", func() {
				Expect(testUI.Out).To(Say("Creating an app manifest from current settings of app some-app in org some-org / space some-space as some-user..."))
				Expect(testUI.Err).To(Say("some-warning"))
				Expect(executeErr).To(MatchError("some-error"))
			})
		})

		When("creating the manifest succeeds", func() {
			BeforeEach(func() {
				fakeActor.CreateApplicationManifestByNameAndSpaceReturns(manifest.Application{}, v2v3action.Warnings{"some-warning"}, nil)
			})

			It("displays the file it created and returns no errors", func() {
				Expect(testUI.Out).To(Say("Creating an app manifest from current settings of app some-app in org some-org / space some-space as some-user..."))
				Expect(testUI.Err).To(Say("some-warning"))
				Expect(testUI.Out).To(Say("OK"))
				Expect(testUI.Out).To(Say("Manifest file created successfully at some-file-path"))
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeActor.CreateApplicationManifestByNameAndSpaceCallCount()).To(Equal(1))
				appArg, spaceArg := fakeActor.CreateApplicationManifestByNameAndSpaceArgsForCall(0)
				Expect(appArg).To(Equal("some-app"))
				Expect(spaceArg).To(Equal("some-space-guid"))

				Expect(fakeActor.WriteApplicationManifestCallCount()).To(Equal(1))
				manifestArg, pathArg := fakeActor.WriteApplicationManifestArgsForCall(0)
				Expect(pathArg).To(Equal("some-file-path"))
				Expect(manifestArg).To(Equal(manifest.Application{}))
			})

			When("no filepath is provided", func() {
				BeforeEach(func() {
					cmd.FilePath = ""
				})

				It("creates application manifest in current directry as <app-name>-manifest.yml", func() {
					Expect(testUI.Out).To(Say("Creating an app manifest from current settings of app some-app in org some-org / space some-space as some-user..."))
					Expect(testUI.Err).To(Say("some-warning"))
					Expect(testUI.Out).To(Say("OK"))
					Expect(testUI.Out).To(Say("Manifest file created successfully at .+some-app_manifest\\.yml"))
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(fakeActor.CreateApplicationManifestByNameAndSpaceCallCount()).To(Equal(1))
					appArg, spaceArg := fakeActor.CreateApplicationManifestByNameAndSpaceArgsForCall(0)
					Expect(appArg).To(Equal("some-app"))
					Expect(spaceArg).To(Equal("some-space-guid"))

					Expect(fakeActor.WriteApplicationManifestCallCount()).To(Equal(1))
					manifestArg, pathArg := fakeActor.WriteApplicationManifestArgsForCall(0)
					Expect(pathArg).To(Equal(filepath.FromSlash("./some-app_manifest.yml")))
					Expect(manifestArg).To(Equal(manifest.Application{}))
				})
			})
		})
	})
})
