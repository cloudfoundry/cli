package v7_test

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
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

var _ = Describe("create-app-manifest Command", func() {
	var (
		cmd             CreateAppManifestCommand
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

		cmd = CreateAppManifestCommand{
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
			fakeActor.GetCurrentUserReturns(
				configv3.User{Name: "some-user"},
				nil)
		})

		When("creating the manifest errors", func() {
			BeforeEach(func() {
				fakeActor.GetRawApplicationManifestByNameAndSpaceReturns(nil, v7action.Warnings{"some-warning"}, errors.New("some-error"))
			})

			It("returns the error, prints warnings", func() {
				Expect(testUI.Out).To(Say("Creating an app manifest from current settings of app some-app in org some-org / space some-space as some-user..."))
				Expect(testUI.Err).To(Say("some-warning"))
				Expect(executeErr).To(MatchError("some-error"))
			})
		})

		When("creating the manifest succeeds", func() {
			var tempDir string
			var yamlContents string
			var pathToYAMLFile string

			BeforeEach(func() {
				var err error
				tempDir, err = ioutil.TempDir("", "create-app-manifest-unit")
				Expect(err).ToNot(HaveOccurred())
				cmd.PWD = tempDir

				yamlContents = `---\n- banana`
				fakeActor.GetRawApplicationManifestByNameAndSpaceReturns([]byte(yamlContents), v7action.Warnings{"some-warning"}, nil)
				pathToYAMLFile = filepath.Join(tempDir, "some-app_manifest.yml")
			})

			AfterEach(func() {
				Expect(os.RemoveAll(tempDir)).ToNot(HaveOccurred())
			})

			It("creates application manifest in current directory as <app-name>-manifest.yml", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeActor.GetRawApplicationManifestByNameAndSpaceCallCount()).To(Equal(1))
				appArg, spaceArg := fakeActor.GetRawApplicationManifestByNameAndSpaceArgsForCall(0)
				Expect(appArg).To(Equal("some-app"))
				Expect(spaceArg).To(Equal("some-space-guid"))

				fileContents, err := ioutil.ReadFile(pathToYAMLFile)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(fileContents)).To(Equal(yamlContents))
			})

			It("displays the file it created and returns no errors", func() {
				Expect(testUI.Out).To(Say("Creating an app manifest from current settings of app some-app in org some-org / space some-space as some-user..."))
				Expect(testUI.Err).To(Say("some-warning"))
				Expect(testUI.Out).To(Say("Manifest file created successfully at %s", regexp.QuoteMeta(filepath.Join(tempDir, "some-app_manifest.yml"))))
				Expect(testUI.Out).To(Say("OK"))
				Expect(executeErr).ToNot(HaveOccurred())
			})

			When("a filepath is provided", func() {
				var flagPath string

				BeforeEach(func() {
					flagPath = filepath.Join(tempDir, "my-special-manifest.yml")
					cmd.FilePath = flag.Path(flagPath)
				})

				It("creates application manifest at the specified location", func() {
					Expect(testUI.Out).To(Say("Creating an app manifest from current settings of app some-app in org some-org / space some-space as some-user..."))
					Expect(testUI.Err).To(Say("some-warning"))
					Expect(testUI.Out).To(Say("Manifest file created successfully at %s", regexp.QuoteMeta(flagPath)))
					Expect(testUI.Out).To(Say("OK"))
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(fakeActor.GetRawApplicationManifestByNameAndSpaceCallCount()).To(Equal(1))
					appArg, spaceArg := fakeActor.GetRawApplicationManifestByNameAndSpaceArgsForCall(0)
					Expect(appArg).To(Equal("some-app"))
					Expect(spaceArg).To(Equal("some-space-guid"))

					fileContents, err := ioutil.ReadFile(flagPath)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(fileContents)).To(Equal(yamlContents))
				})
			})
		})

		When("writing the file errors", func() {
			var yamlContents string
			BeforeEach(func() {
				cmd.PWD = filepath.Join("should", "be", "unwritable")

				yamlContents = `---\n- banana`
				fakeActor.GetRawApplicationManifestByNameAndSpaceReturns([]byte(yamlContents), v7action.Warnings{"some-warning"}, nil)
			})

			It("returns a 'FileCreationError' error", func() {
				Expect(executeErr.Error()).To(ContainSubstring("Error creating file:"))
			})
		})
	})
})
