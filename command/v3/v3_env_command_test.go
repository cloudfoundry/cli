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

var _ = Describe("v3-env Command", func() {
	var (
		cmd             v3.V3EnvCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v3fakes.FakeV3EnvActor
		binaryName      string
		executeErr      error
		appName         string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v3fakes.FakeV3EnvActor)

		cmd = v3.V3EnvCommand{
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

			Context("when getting the environment returns env vars for all groups", func() {
				BeforeEach(func() {
					envGroups := v3action.EnvironmentVariableGroups{
						SystemProvided:      map[string]interface{}{"system-name": map[string]interface{}{"mysql": []string{"system-value"}}},
						ApplicationProvided: map[string]interface{}{"application-name": "application-value"},
						UserProvided:        map[string]interface{}{"user-name": "user-value"},
						RunningGroup:        map[string]interface{}{"running-name": "running-value"},
						StagingGroup:        map[string]interface{}{"staging-name": "staging-value"},
					}
					fakeActor.GetEnvironmentVariablesByApplicationNameAndSpaceReturns(envGroups, v3action.Warnings{"get-warning-1", "get-warning-2"}, nil)
				})

				It("displays the environment variable and value pair", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say("Getting env variables for app some-app in org some-org / space some-space as banana\\.\\.\\."))
					Expect(testUI.Out).To(Say("System-Provided:"))
					Expect(testUI.Out).To(Say("system-name: {"))
					Expect(testUI.Out).To(Say(`"mysql": \[`))
					Expect(testUI.Out).To(Say(`"system-value"`))
					Expect(testUI.Out).To(Say("\\]"))
					Expect(testUI.Out).To(Say("}"))
					Expect(testUI.Out).To(Say(`application-name: "application-value"`))

					Expect(testUI.Out).To(Say("User-Provided:"))
					Expect(testUI.Out).To(Say(`user-name: "user-value"`))

					Expect(testUI.Out).To(Say("Running Environment Variable Groups:"))
					Expect(testUI.Out).To(Say(`running-name: "running-value"`))

					Expect(testUI.Out).To(Say("Staging Environment Variable Groups:"))
					Expect(testUI.Out).To(Say(`staging-name: "staging-value"`))

					Expect(testUI.Err).To(Say("get-warning-1"))
					Expect(testUI.Err).To(Say("get-warning-2"))

					Expect(fakeActor.GetEnvironmentVariablesByApplicationNameAndSpaceCallCount()).To(Equal(1))
					appName, spaceGUID := fakeActor.GetEnvironmentVariablesByApplicationNameAndSpaceArgsForCall(0)
					Expect(appName).To(Equal("some-app"))
					Expect(spaceGUID).To(Equal("some-space-guid"))
				})
			})

			Context("when getting the environment returns empty env vars for all groups", func() {
				BeforeEach(func() {
					envGroups := v3action.EnvironmentVariableGroups{
						SystemProvided:      map[string]interface{}{},
						ApplicationProvided: map[string]interface{}{},
						UserProvided:        map[string]interface{}{},
						RunningGroup:        map[string]interface{}{},
						StagingGroup:        map[string]interface{}{},
					}
					fakeActor.GetEnvironmentVariablesByApplicationNameAndSpaceReturns(envGroups, v3action.Warnings{"get-warning-1", "get-warning-2"}, nil)
				})

				It("displays helpful messages", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say("Getting env variables for app some-app in org some-org / space some-space as banana\\.\\.\\."))
					Expect(testUI.Out).To(Say("No system-provided env variables have been set"))

					Expect(testUI.Out).To(Say("No user-provided env variables have been set"))

					Expect(testUI.Out).To(Say("No running env variables have been set"))

					Expect(testUI.Out).To(Say("No staging env variables have been set"))

					Expect(testUI.Err).To(Say("get-warning-1"))
					Expect(testUI.Err).To(Say("get-warning-2"))

					Expect(fakeActor.GetEnvironmentVariablesByApplicationNameAndSpaceCallCount()).To(Equal(1))
					appName, spaceGUID := fakeActor.GetEnvironmentVariablesByApplicationNameAndSpaceArgsForCall(0)
					Expect(appName).To(Equal("some-app"))
					Expect(spaceGUID).To(Equal("some-space-guid"))
				})
			})

			Context("when the get environment variables returns an unknown error", func() {
				var expectedErr error
				BeforeEach(func() {
					expectedErr = errors.New("some-error")
					fakeActor.GetEnvironmentVariablesByApplicationNameAndSpaceReturns(v3action.EnvironmentVariableGroups{}, v3action.Warnings{"get-warning-1", "get-warning-2"}, expectedErr)
				})

				It("returns the error", func() {
					Expect(executeErr).To(Equal(expectedErr))
					Expect(testUI.Out).To(Say("Getting env variables for app some-app in org some-org / space some-space as banana\\.\\.\\."))

					Expect(testUI.Err).To(Say("get-warning-1"))
					Expect(testUI.Err).To(Say("get-warning-2"))
				})
			})
		})
	})
})
