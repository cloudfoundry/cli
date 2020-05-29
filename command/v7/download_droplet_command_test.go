package v7_test

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/translatableerror"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("download-droplet Command", func() {

	var (
		cmd             DownloadDropletCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		cmd = DownloadDropletCommand{
			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		cmd.RequiredArgs.AppName = "some-app"

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
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

	When("downloading the droplet succeeds", func() {
		var pathToDropletFile string

		BeforeEach(func() {
			fakeActor.DownloadDropletByAppNameReturns([]byte("some-droplet"), "some-droplet-guid", v7action.Warnings{"some-warning"}, nil)

			pathToDropletFile = filepath.Join("droplet_some-droplet-guid.tgz")
		})

		AfterEach(func() {
			Expect(os.Remove("droplet_some-droplet-guid.tgz")).ToNot(HaveOccurred())
		})

		It("creates a droplet tarball in the current directory", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(fakeActor.DownloadDropletByAppNameCallCount()).To(Equal(1))
			appArg, spaceGUIDArg := fakeActor.DownloadDropletByAppNameArgsForCall(0)
			Expect(appArg).To(Equal("some-app"))
			Expect(spaceGUIDArg).To(Equal("some-space-guid"))

			fileContents, err := ioutil.ReadFile(pathToDropletFile)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(fileContents)).To(Equal("some-droplet"))
		})

		It("displays the file it created and returns no errors", func() {
			Expect(testUI.Out).To(Say("Downloading current droplet for app some-app in org some-org / space some-space as some-user..."))
			Expect(testUI.Err).To(Say("some-warning"))
			Expect(testUI.Out).To(Say(`Droplet downloaded successfully at .*droplet_some-droplet-guid.tgz`))
			Expect(testUI.Out).To(Say("OK"))
			Expect(executeErr).ToNot(HaveOccurred())
		})
	})

	When("there is an error downloading the droplet", func() {
		BeforeEach(func() {
			fakeActor.DownloadDropletByAppNameReturns([]byte{}, "", v7action.Warnings{"some-warning"}, errors.New("something went wrong"))
		})

		It("displays warnings and returns a 'NoCurrentDropletForAppError'", func() {
			Expect(testUI.Err).To(Say("some-warning"))
			Expect(executeErr).To(MatchError("something went wrong"))
		})
	})

	When("the app does not have a current droplet", func() {
		BeforeEach(func() {
			fakeActor.DownloadDropletByAppNameReturns([]byte{}, "", v7action.Warnings{"some-warning"}, actionerror.DropletNotFoundError{})
		})

		It("displays warnings and returns a 'NoCurrentDropletForAppError'", func() {
			Expect(testUI.Err).To(Say("some-warning"))
			Expect(executeErr).To(MatchError(translatableerror.NoCurrentDropletForAppError{AppName: "some-app"}))
		})
	})

	When("writing the droplet file fails", func() {
		BeforeEach(func() {
			cmd.CWD = filepath.Join("should", "be", "unwritable")
			fakeActor.DownloadDropletByAppNameReturns([]byte("some-droplet"), "some-droplet-guid", v7action.Warnings{"some-warning"}, nil)
		})

		It("returns a 'DropletFileError' error", func() {
			Expect(testUI.Err).To(Say("some-warning"))
			Expect(executeErr.Error()).To(ContainSubstring("Error creating droplet file:"))
		})
	})
})
