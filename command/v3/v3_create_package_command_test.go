package v3_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v3"
	"code.cloudfoundry.org/cli/command/v3/shared"
	"code.cloudfoundry.org/cli/command/v3/v3fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("v3-create-package Command", func() {
	var (
		cmd             v3.V3CreatePackageCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v3fakes.FakeV3CreatePackageActor
		binaryName      string
		executeErr      error
		app             string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v3fakes.FakeV3CreatePackageActor)

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		app = "some-app"

		packageDisplayer := shared.NewPackageDisplayer(
			testUI,
			fakeConfig,
		)

		cmd = v3.V3CreatePackageCommand{
			UI:               testUI,
			Config:           fakeConfig,
			SharedActor:      fakeSharedActor,
			Actor:            fakeActor,
			RequiredArgs:     flag.AppName{AppName: app},
			PackageDisplayer: packageDisplayer,
		}

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
			fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: binaryName}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeTrue())
			Expect(checkTargetedSpace).To(BeTrue())
		})
	})

	Context("when the user is logged in", func() {
		BeforeEach(func() {
			fakeConfig.CurrentUserReturns(configv3.User{Name: "banana"}, nil)
			fakeConfig.TargetedSpaceReturns(configv3.Space{Name: "some-space", GUID: "some-space-guid"})
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "some-org"})
		})

		Context("when no flags are set", func() {
			Context("when the create is successful", func() {
				BeforeEach(func() {
					myPackage := v3action.Package{GUID: "1234"}
					fakeActor.CreateAndUploadBitsPackageByApplicationNameAndSpaceReturns(myPackage, v3action.Warnings{"I am a warning", "I am also a warning"}, nil)
				})

				It("displays the header and ok", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say("Uploading and creating bits package for app some-app in org some-org / space some-space as banana..."))
					Expect(testUI.Out).To(Say("package guid: 1234"))
					Expect(testUI.Out).To(Say("OK"))

					Expect(testUI.Err).To(Say("I am a warning"))
					Expect(testUI.Err).To(Say("I am also a warning"))

					Expect(fakeActor.CreateAndUploadBitsPackageByApplicationNameAndSpaceCallCount()).To(Equal(1))

					appName, spaceGUID, bitsPath := fakeActor.CreateAndUploadBitsPackageByApplicationNameAndSpaceArgsForCall(0)
					Expect(appName).To(Equal(app))
					Expect(spaceGUID).To(Equal("some-space-guid"))
					Expect(bitsPath).To(BeEmpty())
				})
			})

			Context("when the create is unsuccessful", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("I am an error")
					fakeActor.CreateAndUploadBitsPackageByApplicationNameAndSpaceReturns(v3action.Package{}, v3action.Warnings{"I am a warning", "I am also a warning"}, expectedErr)
				})

				It("displays the header and error", func() {
					Expect(executeErr).To(MatchError(expectedErr))

					Expect(testUI.Out).To(Say("Uploading and creating bits package for app some-app in org some-org / space some-space as banana..."))

					Expect(testUI.Err).To(Say("I am a warning"))
					Expect(testUI.Err).To(Say("I am also a warning"))
				})
			})
		})

		Context("when the docker image is provided", func() {
			BeforeEach(func() {
				cmd.DockerImage.Path = "some-docker-image"
				fakeActor.CreateDockerPackageByApplicationNameAndSpaceReturns(v3action.Package{GUID: "1234"}, v3action.Warnings{"I am a warning", "I am also a warning"}, nil)
			})

			It("creates the docker package", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Out).To(Say("Creating docker package for app some-app in org some-org / space some-space as banana..."))
				Expect(testUI.Out).To(Say("package guid: 1234"))
				Expect(testUI.Out).To(Say("OK"))

				Expect(testUI.Err).To(Say("I am a warning"))
				Expect(testUI.Err).To(Say("I am also a warning"))

				Expect(fakeActor.CreateDockerPackageByApplicationNameAndSpaceCallCount()).To(Equal(1))

				appName, spaceGUID, dockerImageCredentials := fakeActor.CreateDockerPackageByApplicationNameAndSpaceArgsForCall(0)
				Expect(appName).To(Equal(app))
				Expect(spaceGUID).To(Equal("some-space-guid"))
				Expect(dockerImageCredentials.Path).To(Equal("some-docker-image"))
			})
		})

		Context("when the path is provided", func() {
			BeforeEach(func() {
				cmd.AppPath = "some-app-path"
				fakeActor.CreateAndUploadBitsPackageByApplicationNameAndSpaceReturns(v3action.Package{GUID: "1234"}, v3action.Warnings{"I am a warning", "I am also a warning"}, nil)
			})

			It("creates the package using the specified directory", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Out).To(Say("Uploading and creating bits package for app some-app in org some-org / space some-space as banana..."))

				Expect(testUI.Err).To(Say("I am a warning"))
				Expect(testUI.Err).To(Say("I am also a warning"))

				Expect(fakeActor.CreateAndUploadBitsPackageByApplicationNameAndSpaceCallCount()).To(Equal(1))

				appName, spaceGUID, appPath := fakeActor.CreateAndUploadBitsPackageByApplicationNameAndSpaceArgsForCall(0)
				Expect(appName).To(Equal(app))
				Expect(spaceGUID).To(Equal("some-space-guid"))
				Expect(appPath).To(Equal("some-app-path"))
			})
		})

		Context("when the docker image and path are both provided", func() {
			BeforeEach(func() {
				cmd.AppPath = "some-app-path"
				cmd.DockerImage.Path = "some-docker-image"
			})

			It("displays an argument combination error", func() {
				argumentCombinationError := translatableerror.ArgumentCombinationError{
					Args: []string{"--docker-image", "-o", "-p"},
				}

				Expect(executeErr).To(MatchError(argumentCombinationError))
			})
		})
	})
})
