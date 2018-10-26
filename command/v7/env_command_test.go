package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("env Command", func() {
	var (
		cmd             EnvCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeEnvActor
		binaryName      string
		executeErr      error
		appName         string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeEnvActor)

		cmd = EnvCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		appName = "some-app"

		cmd.RequiredArgs.AppName = appName
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("checking target fails", func() {
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

	When("the user is logged in, an org is targeted and a space is targeted", func() {
		BeforeEach(func() {
			fakeConfig.TargetedSpaceReturns(configv3.Space{Name: "some-space", GUID: "some-space-guid"})
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "some-org"})
		})

		When("getting the current user returns an error", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserReturns(configv3.User{}, errors.New("some-error"))
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError("some-error"))
			})
		})

		When("getting the current user succeeds", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserReturns(configv3.User{Name: "banana"}, nil)
			})

			When("getting the environment returns env vars for all groups", func() {
				BeforeEach(func() {
					envGroups := v7action.EnvironmentVariableGroups{
						System:               map[string]interface{}{"system-name": map[string]interface{}{"mysql": []string{"system-value"}}},
						Application:          map[string]interface{}{"application-name": "application-value"},
						EnvironmentVariables: map[string]interface{}{"user-name": "user-value"},
						Running:              map[string]interface{}{"running-name": "running-value"},
						Staging:              map[string]interface{}{"staging-name": "staging-value"},
					}
					fakeActor.GetEnvironmentVariablesByApplicationNameAndSpaceReturns(envGroups, v7action.Warnings{"get-warning-1", "get-warning-2"}, nil)
				})

				It("displays the environment variable and value pair", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say("Getting env variables for app some-app in org some-org / space some-space as banana\\.\\.\\."))
					Expect(testUI.Out).To(Say("OK"))
					Expect(testUI.Out).To(Say("System-Provided:"))
					Expect(testUI.Out).To(Say("system-name: {"))
					Expect(testUI.Out).To(Say(`"mysql": \[`))
					Expect(testUI.Out).To(Say(`"system-value"`))
					Expect(testUI.Out).To(Say("\\]"))
					Expect(testUI.Out).To(Say("}"))
					Expect(testUI.Out).To(Say(`application-name: "application-value"`))

					Expect(testUI.Out).To(Say("User-Provided:"))
					Expect(testUI.Out).To(Say(`user-name: user-value`))

					Expect(testUI.Out).To(Say("Running Environment Variable Groups:"))
					Expect(testUI.Out).To(Say(`running-name: running-value`))

					Expect(testUI.Out).To(Say("Staging Environment Variable Groups:"))
					Expect(testUI.Out).To(Say(`staging-name: staging-value`))

					Expect(testUI.Err).To(Say("get-warning-1"))
					Expect(testUI.Err).To(Say("get-warning-2"))

					Expect(fakeActor.GetEnvironmentVariablesByApplicationNameAndSpaceCallCount()).To(Equal(1))
					appName, spaceGUID := fakeActor.GetEnvironmentVariablesByApplicationNameAndSpaceArgsForCall(0)
					Expect(appName).To(Equal("some-app"))
					Expect(spaceGUID).To(Equal("some-space-guid"))
				})

				Describe("sorting of non-json environment variables", func() {
					BeforeEach(func() {
						envGroups := v7action.EnvironmentVariableGroups{
							System:      map[string]interface{}{},
							Application: map[string]interface{}{},
							EnvironmentVariables: map[string]interface{}{
								"alpha":   "1",
								"charlie": "1",
								"bravo":   "1",
							},
							Running: map[string]interface{}{
								"foxtrot": "1",
								"delta":   "1",
								"echo":    "1",
							},
							Staging: map[string]interface{}{
								"hotel": "1",
								"india": "1",
								"golf":  "1",
							},
						}
						fakeActor.GetEnvironmentVariablesByApplicationNameAndSpaceReturns(envGroups, v7action.Warnings{"get-warning-1", "get-warning-2"}, nil)
					})

					It("sorts the EnvironmentVariables alphabetically", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say(`alpha: 1`))
						Expect(testUI.Out).To(Say(`bravo: 1`))
						Expect(testUI.Out).To(Say(`charlie: 1`))
					})

					It("sorts the Running alphabetically", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say(`delta: 1`))
						Expect(testUI.Out).To(Say(`echo: 1`))
						Expect(testUI.Out).To(Say(`foxtrot: 1`))
					})

					It("sorts the Staging alphabetically", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say(`golf: 1`))
						Expect(testUI.Out).To(Say(`hotel: 1`))
						Expect(testUI.Out).To(Say(`india: 1`))
					})
				})
			})

			When("getting the environment returns empty env vars for all groups", func() {
				BeforeEach(func() {
					envGroups := v7action.EnvironmentVariableGroups{
						System:               map[string]interface{}{},
						Application:          map[string]interface{}{},
						EnvironmentVariables: map[string]interface{}{},
						Running:              map[string]interface{}{},
						Staging:              map[string]interface{}{},
					}
					fakeActor.GetEnvironmentVariablesByApplicationNameAndSpaceReturns(envGroups, v7action.Warnings{"get-warning-1", "get-warning-2"}, nil)
				})

				It("displays helpful messages", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say("Getting env variables for app some-app in org some-org / space some-space as banana\\.\\.\\."))
					Expect(testUI.Out).To(Say("OK"))

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

			When("the get environment variables returns an unknown error", func() {
				var expectedErr error
				BeforeEach(func() {
					expectedErr = errors.New("some-error")
					fakeActor.GetEnvironmentVariablesByApplicationNameAndSpaceReturns(v7action.EnvironmentVariableGroups{}, v7action.Warnings{"get-warning-1", "get-warning-2"}, expectedErr)
				})

				It("returns the error", func() {
					Expect(executeErr).To(Equal(expectedErr))
					Expect(testUI.Out).To(Say("Getting env variables for app some-app in org some-org / space some-space as banana\\.\\.\\."))
					Expect(testUI.Out).To(Say("OK"))

					Expect(testUI.Err).To(Say("get-warning-1"))
					Expect(testUI.Err).To(Say("get-warning-2"))
				})
			})
		})
	})
})
