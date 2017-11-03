package v3_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v3"
	"code.cloudfoundry.org/cli/command/v3/v3fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("v3-unset-env Command", func() {
	var (
		cmd             v3.V3UnsetEnvCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v3fakes.FakeV3UnsetEnvActor
		binaryName      string
		executeErr      error
		appName         string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v3fakes.FakeV3UnsetEnvActor)

		cmd = v3.V3UnsetEnvCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

		fakeActor.CloudControllerAPIVersionReturns(ccversion.MinVersionV3)
		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		appName = "some-app"

		cmd.RequiredArgs.AppName = appName
		cmd.RequiredArgs.EnvironmentVariableName = "some-key"
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

	Context("when the user is logged in, an org is targeted and a space is targeted", func() {
		BeforeEach(func() {
			fakeConfig.TargetedSpaceReturns(configv3.Space{Name: "some-space", GUID: "some-space-guid"})
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "some-org"})
		})

		Context("when getting the current user returns an error", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserReturns(configv3.User{}, errors.New("some-error"))
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError("some-error"))
			})
		})

		Context("when getting the current user succeeds", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserReturns(configv3.User{Name: "banana"}, nil)
			})

			Context("when unsetting the environment variable succeeds", func() {
				BeforeEach(func() {
					fakeActor.UnsetEnvironmentVariableByApplicationNameAndSpaceReturns(v3action.Warnings{"set-warning-1", "set-warning-2"}, nil)
				})

				It("sets the environment variable and value pair", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say("Removing env variable some-key from app some-app in org some-org / space some-space as banana\\.\\.\\."))

					Expect(testUI.Err).To(Say("set-warning-1"))
					Expect(testUI.Err).To(Say("set-warning-2"))
					Expect(testUI.Out).To(Say("OK"))
					Expect(testUI.Out).To(Say("TIP: Use 'cf v3-stage some-app' to ensure your env variable changes take effect\\."))

					Expect(fakeActor.UnsetEnvironmentVariableByApplicationNameAndSpaceCallCount()).To(Equal(1))
					appName, spaceGUID, envVarName := fakeActor.UnsetEnvironmentVariableByApplicationNameAndSpaceArgsForCall(0)
					Expect(appName).To(Equal("some-app"))
					Expect(spaceGUID).To(Equal("some-space-guid"))
					Expect(envVarName).To(Equal("some-key"))
				})
			})

			Context("when unsetting the environment returns an EnvironmentVariableNotSetError", func() {
				BeforeEach(func() {
					fakeActor.UnsetEnvironmentVariableByApplicationNameAndSpaceReturns(v3action.Warnings{"unset-warning-1", "unset-warning-2"}, actionerror.EnvironmentVariableNotSetError{EnvironmentVariableName: "some-key"})
				})

				It("displays okay and the error", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say("Removing env variable some-key from app some-app in org some-org / space some-space as banana\\.\\.\\."))
					Expect(testUI.Out).To(Say("Env variable some-key was not set"))
					Expect(testUI.Out).To(Say("OK"))
				})
			})

			Context("when the set environment variable returns an unknown error", func() {
				var expectedErr error
				BeforeEach(func() {
					expectedErr = errors.New("some-error")
					fakeActor.UnsetEnvironmentVariableByApplicationNameAndSpaceReturns(v3action.Warnings{"get-warning-1", "get-warning-2"}, expectedErr)
				})

				It("returns the error", func() {
					Expect(executeErr).To(Equal(expectedErr))
					Expect(testUI.Out).To(Say("Removing env variable some-key from app some-app in org some-org / space some-space as banana\\.\\.\\."))

					Expect(testUI.Err).To(Say("get-warning-1"))
					Expect(testUI.Err).To(Say("get-warning-2"))
					Expect(testUI.Out).ToNot(Say("OK"))
				})
			})
		})
	})
})
