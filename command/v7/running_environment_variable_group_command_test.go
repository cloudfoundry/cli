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

var _ = Describe("running-environment-variable-group Command", func() {
	var (
		cmd             RunningEnvironmentVariableGroupCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		executeErr      error
		args            []string
		binaryName      string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)
		args = nil

		cmd = RunningEnvironmentVariableGroupCommand{
			BaseCommand: command.BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(args)
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
			fakeConfig.CurrentUserReturns(configv3.User{Name: "apple"}, nil)
		})

		It("should print text indicating its running", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(testUI.Out).To(Say(`Getting the running environment variable group as apple\.\.\.`))
		})

		When("getting the environment variables fails", func() {
			BeforeEach(func() {
				fakeActor.GetEnvironmentVariableGroupReturns(
					nil,
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

		When("getting the environment variables succeeds", func() {
			When("there are some environment variables", func() {
				BeforeEach(func() {
					envVars := v7action.EnvironmentVariableGroup{
						"key_one": {IsSet: true, Value: "value_one"},
						"key_two": {IsSet: true, Value: "value_two"},
					}

					fakeActor.GetEnvironmentVariableGroupReturns(
						envVars,
						v7action.Warnings{"some-warning-1", "some-warning-2"},
						nil,
					)
				})

				It("prints a table of env vars", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(testUI.Err).To(Say("some-warning-1"))
					Expect(testUI.Err).To(Say("some-warning-2"))
					Expect(testUI.Out).To(Say(`variable name\s+assigned value`))
					// We have to do complex regex match, for the results are returned in a random order
					kv1Regex := `key_one\s+value_one`
					kv2Regex := `key_two\s+value_two`
					Expect(testUI.Out).To(Say("(%s\n%s)|(%s\n%s)", kv1Regex, kv2Regex, kv2Regex, kv1Regex))
				})
			})

			When("there are no environment variables in the group", func() {
				BeforeEach(func() {
					envVars := v7action.EnvironmentVariableGroup{}

					fakeActor.GetEnvironmentVariableGroupReturns(
						envVars,
						v7action.Warnings{"some-warning-1", "some-warning-2"},
						nil,
					)
				})

				It("prints a message indicating empty group", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(testUI.Err).To(Say("some-warning-1"))
					Expect(testUI.Err).To(Say("some-warning-2"))
					Expect(testUI.Out).To(Say("No running environment variable group has been set."))
				})
			})
		})
	})
})
