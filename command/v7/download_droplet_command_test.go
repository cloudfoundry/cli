package v7_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/translatableerror"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo/v2"
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
		fakeActor.GetCurrentUserReturns(
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
		var (
			pathToDropletFile string
			dropletGUID       string
		)

		BeforeEach(func() {
			dropletGUID = RandomString("fake-droplet-guid")
			fakeActor.DownloadCurrentDropletByAppNameReturns([]byte("some-droplet-bytes"), dropletGUID, v7action.Warnings{"some-warning"}, nil)

			currentDir, _ := os.Getwd()
			pathToDropletFile = filepath.Join(currentDir, fmt.Sprintf("droplet_%s.tgz", dropletGUID))
		})

		AfterEach(func() {
			Expect(os.Remove(pathToDropletFile)).ToNot(HaveOccurred())
		})

		It("creates a droplet tarball in the current directory", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(fakeActor.DownloadCurrentDropletByAppNameCallCount()).To(Equal(1))
			appArg, spaceGUIDArg := fakeActor.DownloadCurrentDropletByAppNameArgsForCall(0)
			Expect(appArg).To(Equal("some-app"))
			Expect(spaceGUIDArg).To(Equal("some-space-guid"))

			fileContents, err := ioutil.ReadFile(pathToDropletFile)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(fileContents)).To(Equal("some-droplet-bytes"))
		})

		It("displays the file it created and returns no errors", func() {
			Expect(testUI.Out).To(Say("Downloading current droplet for app some-app in org some-org / space some-space as some-user..."))
			Expect(testUI.Err).To(Say("some-warning"))
			Expect(testUI.Out).To(Say(`Droplet downloaded successfully at .*droplet_%s.tgz`, dropletGUID))
			Expect(testUI.Out).To(Say("OK"))
			Expect(executeErr).ToNot(HaveOccurred())
		})
	})

	When("the droplet guid is passed in", func() {
		var (
			dropletGUID       string
			pathToDropletFile string
		)

		BeforeEach(func() {
			dropletGUID = RandomString("fake-droplet-guid")
			pathToDropletFile = fmt.Sprintf("droplet_%s.tgz", dropletGUID)

			setFlag(&cmd, "--droplet", dropletGUID)

			fakeActor.DownloadDropletByGUIDAndAppNameReturns([]byte("some-droplet-bytes"), v7action.Warnings{"some-warning"}, nil)
		})

		AfterEach(func() {
			Expect(os.Remove(pathToDropletFile)).ToNot(HaveOccurred())
		})

		It("creates a droplet tarball in the current directory", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(fakeActor.DownloadDropletByGUIDAndAppNameCallCount()).To(Equal(1))
			dropletGUIDArg, appArg, spaceGUIDArg := fakeActor.DownloadDropletByGUIDAndAppNameArgsForCall(0)
			Expect(dropletGUIDArg).To(Equal(dropletGUID))
			Expect(appArg).To(Equal("some-app"))
			Expect(spaceGUIDArg).To(Equal("some-space-guid"))

			fileContents, err := ioutil.ReadFile(pathToDropletFile)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(fileContents)).To(Equal("some-droplet-bytes"))
		})

		It("displays the file it created and returns no errors", func() {
			Expect(testUI.Out).To(Say("Downloading droplet %s for app some-app in org some-org / space some-space as some-user...", dropletGUID))
			Expect(testUI.Err).To(Say("some-warning"))
			Expect(testUI.Out).To(Say(`Droplet downloaded successfully at .*droplet_%s.tgz`, dropletGUID))
			Expect(testUI.Out).To(Say("OK"))
			Expect(executeErr).ToNot(HaveOccurred())
		})
	})

	When("a path to a file is passed in", func() {
		var filePath string
		BeforeEach(func() {
			filePath = RandomString("fake-file")

			setFlag(&cmd, "--path", filePath)
			fakeActor.DownloadCurrentDropletByAppNameReturns([]byte("some-droplet"), "some-droplet-guid", v7action.Warnings{"some-warning"}, nil)
		})

		AfterEach(func() {
			Expect(os.Remove(filePath)).ToNot(HaveOccurred())
		})

		It("creates a droplet tarball at the specified path", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			fileContents, err := ioutil.ReadFile(filePath)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(fileContents)).To(Equal("some-droplet"))
		})

		It("displays the file it created and returns no errors", func() {
			Expect(testUI.Out).To(Say("Downloading current droplet for app some-app in org some-org / space some-space as some-user..."))
			Expect(testUI.Err).To(Say("some-warning"))
			Expect(testUI.Out).To(Say(`Droplet downloaded successfully at %s`, filePath))
			Expect(testUI.Out).To(Say("OK"))
			Expect(executeErr).ToNot(HaveOccurred())
		})
	})

	When("a path to an existing directory is passed in", func() {
		var tmpDir string

		BeforeEach(func() {
			var err error
			tmpDir, err = ioutil.TempDir("", "droplets")
			Expect(err).NotTo(HaveOccurred())

			setFlag(&cmd, "--path", tmpDir)

			fakeActor.DownloadCurrentDropletByAppNameReturns([]byte("some-droplet"), "some-droplet-guid", v7action.Warnings{"some-warning"}, nil)
		})

		AfterEach(func() {
			Expect(os.RemoveAll(tmpDir)).ToNot(HaveOccurred())
		})

		It("creates a droplet tarball at the specified path", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			fileContents, err := ioutil.ReadFile(filepath.Join(tmpDir, "droplet_some-droplet-guid.tgz"))
			Expect(err).ToNot(HaveOccurred())
			Expect(string(fileContents)).To(Equal("some-droplet"))
		})

		It("displays the file it created and returns no errors", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(testUI.Out).To(Say("Downloading current droplet for app some-app in org some-org / space some-space as some-user..."))
			Expect(testUI.Err).To(Say("some-warning"))
			pathRegExp := regexp.QuoteMeta(filepath.Join(tmpDir, "droplet_some-droplet-guid.tgz"))
			Expect(testUI.Out).To(Say(`Droplet downloaded successfully at %s`, pathRegExp))
			Expect(testUI.Out).To(Say("OK"))
		})
	})

	When("a path to a file in an invalid directory is passed in", func() {
		BeforeEach(func() {
			cmd.Path = "not/exist/some-file.tgz"
			fakeActor.DownloadCurrentDropletByAppNameReturns([]byte("some-droplet"), "some-droplet-guid", v7action.Warnings{"some-warning"}, nil)
		})

		It("returns an appropriate error", func() {
			_, ok := executeErr.(translatableerror.DropletFileError)
			Expect(ok).To(BeTrue())
		})
	})

	When("there is an error downloading the droplet", func() {
		BeforeEach(func() {
			fakeActor.DownloadCurrentDropletByAppNameReturns([]byte{}, "", v7action.Warnings{"some-warning"}, errors.New("something went wrong"))
		})

		It("displays warnings and returns an error", func() {
			Expect(testUI.Err).To(Say("some-warning"))
			Expect(executeErr).To(MatchError("something went wrong"))
		})
	})

	When("the app does not have a current droplet", func() {
		BeforeEach(func() {
			fakeActor.DownloadCurrentDropletByAppNameReturns([]byte{}, "", v7action.Warnings{"some-warning"}, actionerror.DropletNotFoundError{})
		})

		It("displays warnings and returns an error", func() {
			Expect(testUI.Err).To(Say("some-warning"))
			Expect(executeErr).To(MatchError(translatableerror.NoDropletForAppError{AppName: "some-app"}))
		})
	})
})
