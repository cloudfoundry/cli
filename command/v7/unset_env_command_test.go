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

var _ = Describe("unset-env Command", func() {
	var (
		cmd             UnsetEnvCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		binaryName      string
		executeErr      error
		appName         string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		cmd = UnsetEnvCommand{
			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		appName = "some-app"

		cmd.RequiredArgs.AppName = appName
		cmd.RequiredArgs.EnvironmentVariableName = "some-key"
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
				fakeActor.GetCurrentUserReturns(configv3.User{}, errors.New("some-error"))
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError("some-error"))
			})
		})

		When("getting the current user succeeds", func() {
			BeforeEach(func() {
				fakeActor.GetCurrentUserReturns(configv3.User{Name: "banana"}, nil)
			})

			When("unsetting the environment variable succeeds", func() {
				BeforeEach(func() {
					fakeActor.UnsetEnvironmentVariableByApplicationNameAndSpaceReturns(v7action.Warnings{"set-warning-1", "set-warning-2"}, nil)
				})

				It("sets the environment variable and value pair", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say(`Removing env variable some-key from app some-app in org some-org / space some-space as banana\.\.\.`))

					Expect(testUI.Err).To(Say("set-warning-1"))
					Expect(testUI.Err).To(Say("set-warning-2"))
					Expect(testUI.Out).To(Say("OK"))
					Expect(testUI.Out).To(Say(`TIP: Use 'cf restage some-app' to ensure your env variable changes take effect\.`))

					Expect(fakeActor.UnsetEnvironmentVariableByApplicationNameAndSpaceCallCount()).To(Equal(1))
					appName, spaceGUID, envVarName := fakeActor.UnsetEnvironmentVariableByApplicationNameAndSpaceArgsForCall(0)
					Expect(appName).To(Equal("some-app"))
					Expect(spaceGUID).To(Equal("some-space-guid"))
					Expect(envVarName).To(Equal("some-key"))
				})
			})

			When("unsetting the environment returns an EnvironmentVariableNotSetError", func() {
				BeforeEach(func() {
					fakeActor.UnsetEnvironmentVariableByApplicationNameAndSpaceReturns(v7action.Warnings{"unset-warning-1", "unset-warning-2"}, actionerror.EnvironmentVariableNotSetError{EnvironmentVariableName: "some-key"})
				})

				It("displays okay and the error", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say(`Removing env variable some-key from app some-app in org some-org / space some-space as banana\.\.\.`))
					Expect(testUI.Out).To(Say("Env variable some-key was not set"))
					Expect(testUI.Out).To(Say("OK"))
				})
			})

			When("the set environment variable returns an unknown error", func() {
				var expectedErr error
				BeforeEach(func() {
					expectedErr = errors.New("some-error")
					fakeActor.UnsetEnvironmentVariableByApplicationNameAndSpaceReturns(v7action.Warnings{"get-warning-1", "get-warning-2"}, expectedErr)
				})

				It("returns the error", func() {
					Expect(executeErr).To(Equal(expectedErr))
					Expect(testUI.Out).To(Say(`Removing env variable some-key from app some-app in org some-org / space some-space as banana\.\.\.`))

					Expect(testUI.Err).To(Say("get-warning-1"))
					Expect(testUI.Err).To(Say("get-warning-2"))
					Expect(testUI.Out).ToNot(Say("OK"))
				})
			})
		})
	})
})
