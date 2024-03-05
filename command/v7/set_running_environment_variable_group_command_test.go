package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/resources"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("set-running-environment-variable-group Command", func() {
	var (
		cmd             SetRunningEnvironmentVariableGroupCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		executeErr      error
		binaryName      string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		cmd = SetRunningEnvironmentVariableGroupCommand{
			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		cmd.RequiredArgs.EnvVarGroupJson = `{"key1":"val1", "key2":"val2"}`

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("the environment is not set up correctly", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: binaryName}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeFalse())
			Expect(checkTargetedSpace).To(BeFalse())
		})
	})

	When("the environment is setup correctly", func() {
		BeforeEach(func() {
			fakeActor.GetCurrentUserReturns(configv3.User{Name: "apple"}, nil)
		})

		It("should print text indicating its running", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(testUI.Out).To(Say(`Setting the contents of the running environment variable group as apple\.\.\.`))
		})

		When("jsonUnmarshalling fails", func() {
			BeforeEach(func() {
				cmd.RequiredArgs.EnvVarGroupJson = "bad json"
			})

			It("should err", func() {
				Expect(executeErr).To(MatchError("Invalid environment variable group provided. Please provide a valid JSON object."))
			})
		})

		When("setting the environment variables fails", func() {
			BeforeEach(func() {
				fakeActor.SetEnvironmentVariableGroupReturns(
					v7action.Warnings{"some-warning-1", "some-warning-2"},
					errors.New("some-error"),
				)
			})

			It("prints warnings and returns error", func() {
				Expect(executeErr).To(MatchError("some-error"))

				Expect(testUI.Err).To(Say("some-warning-1"))
				Expect(testUI.Err).To(Say("some-warning-2"))
			})
		})

		When("setting the environment variables succeeds", func() {
			BeforeEach(func() {
				fakeActor.SetEnvironmentVariableGroupReturns(
					v7action.Warnings{"some-warning-1", "some-warning-2"},
					nil,
				)
			})

			It("should print text indicating its set", func() {
				Expect(fakeActor.SetEnvironmentVariableGroupCallCount()).To(Equal(1))
				group, envVars := fakeActor.SetEnvironmentVariableGroupArgsForCall(0)
				Expect(group).To(Equal(constant.RunningEnvironmentVariableGroup))
				Expect(envVars).To(Equal(resources.EnvironmentVariables{
					"key1": {Value: "val1", IsSet: true},
					"key2": {Value: "val2", IsSet: true},
				}))
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(testUI.Out).To(Say("OK"))
			})
		})
	})
})
