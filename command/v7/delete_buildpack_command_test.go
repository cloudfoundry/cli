package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/v8/actor/actionerror"
	"code.cloudfoundry.org/cli/v8/actor/v7action"
	"code.cloudfoundry.org/cli/v8/command/commandfakes"
	"code.cloudfoundry.org/cli/v8/command/translatableerror"
	. "code.cloudfoundry.org/cli/v8/command/v7"
	"code.cloudfoundry.org/cli/v8/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/v8/util/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("delete-buildpack Command", func() {

	var (
		cmd             DeleteBuildpackCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		input           *Buffer
		binaryName      string
		buildpackName   string
		executeErr      error
	)

	BeforeEach(func() {
		input = NewBuffer()
		fakeActor = new(v7fakes.FakeActor)
		fakeConfig = new(commandfakes.FakeConfig)
		fakeConfig.APIVersionReturns("4.0.0")
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())

		cmd = DeleteBuildpackCommand{
			BaseCommand: BaseCommand{
				Actor:       fakeActor,
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
			},
		}
		binaryName = "faceman"
		buildpackName = "the-buildpack"
		fakeConfig.BinaryNameReturns(binaryName)
		cmd.RequiredArgs.Buildpack = buildpackName
		cmd.Force = true
	})

	When("--lifecyle is provided", func() {
		JustBeforeEach(func() {
			cmd.Lifecycle = "some-lifecycle"
			fakeConfig.APIVersionReturns("3.193.0")
		})
		It("fails when the cc version is below the minimum", func() {
			executeErr = cmd.Execute(nil)

			Expect(executeErr).To(MatchError(translatableerror.MinimumCFAPIVersionNotMetError{
				Command:        "--lifecycle",
				CurrentVersion: "3.193.0",
				MinimumVersion: "3.194.0",
			}))
		})

	})

	When("checking target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: binaryName})
		})

		It("returns an error if the check fails", func() {
			executeErr = cmd.Execute(nil)

			Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: "faceman"}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			shouldCheckTargetedOrg, shouldCheckTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(shouldCheckTargetedOrg).To(BeFalse())
			Expect(shouldCheckTargetedSpace).To(BeFalse())
		})
	})

	When("the DeleteBuildpack actor completes successfully", func() {
		BeforeEach(func() {
			fakeActor.DeleteBuildpackByNameAndStackAndLifecycleReturns(nil, nil)
		})
		JustBeforeEach(func() {
			executeErr = cmd.Execute(nil)
		})

		When("--force is specified", func() {
			BeforeEach(func() {
				cmd.Force = true
			})

			When("a stack is not specified", func() {
				BeforeEach(func() {
					cmd.Stack = ""
				})

				It("prints appropriate output", func() {
					Expect(testUI.Out).To(Say("Deleting buildpack the-buildpack..."))
					Expect(testUI.Out).To(Say("OK"))
				})
			})

			When("a stack is specified", func() {
				BeforeEach(func() {
					cmd.Stack = "a-stack"
				})

				It("prints appropriate output that includes the stack name", func() {
					Expect(testUI.Out).To(Say("Deleting buildpack the-buildpack with stack a-stack..."))
					Expect(testUI.Out).To(Say("OK"))
				})
			})
		})

		When("--force is not specified", func() {
			BeforeEach(func() {
				cmd.Force = false
			})

			When("the user inputs yes", func() {
				BeforeEach(func() {
					_, err := input.Write([]byte("y\n"))
					Expect(err).ToNot(HaveOccurred())
				})

				It("prompted the user for confirmation", func() {
					Expect(testUI.Out).To(Say("Really delete the buildpack the-buildpack?"))
					Expect(testUI.Out).To(Say("Deleting buildpack the-buildpack..."))
					Expect(testUI.Out).To(Say("OK"))
				})
			})

			When("the user inputs no", func() {
				BeforeEach(func() {
					_, err := input.Write([]byte("n\n"))
					Expect(err).ToNot(HaveOccurred())
				})

				It("cancels the delete", func() {
					Expect(testUI.Out).To(Say("Really delete the buildpack the-buildpack?"))
					Expect(testUI.Out).To(Say("Delete cancelled"))
					Expect(testUI.Out).NotTo(Say("Deleting buildpack the-buildpack..."))
				})
			})
		})
	})

	When("the buildpack does not exist", func() {
		When("deleting with a stack", func() {
			BeforeEach(func() {
				fakeActor.DeleteBuildpackByNameAndStackAndLifecycleReturns(v7action.Warnings{"a-warning"}, actionerror.BuildpackNotFoundError{BuildpackName: buildpackName, StackName: "stack!"})
				cmd.Stack = "stack!"
				executeErr = cmd.Execute(nil)
			})

			It("prints warnings and helpful error message (that includes the stack name)", func() {
				Expect(testUI.Err).To(Say("a-warning"))
				Expect(testUI.Err).To(Say(`Buildpack 'the-buildpack' with stack 'stack!' not found\.`))
			})
		})

		When("deleting with a lifecycle", func() {
			BeforeEach(func() {
				fakeActor.DeleteBuildpackByNameAndStackAndLifecycleReturns(v7action.Warnings{"a-warning"}, actionerror.BuildpackNotFoundError{BuildpackName: buildpackName, Lifecycle: "cnb"})
				cmd.Lifecycle = "cnb"
				executeErr = cmd.Execute(nil)
			})

			It("prints warnings and helpful error message (that includes the lifecycle name)", func() {
				Expect(testUI.Err).To(Say("a-warning"))
				Expect(testUI.Err).To(Say(`Buildpack 'the-buildpack' with lifecycle 'cnb' not found\.`))
			})

		})
		When("deleting with a stack and lifecycle", func() {
			BeforeEach(func() {
				fakeActor.DeleteBuildpackByNameAndStackAndLifecycleReturns(v7action.Warnings{"a-warning"}, actionerror.BuildpackNotFoundError{BuildpackName: buildpackName, StackName: "stack!", Lifecycle: "cnb"})
				cmd.Stack = "stack!"
				cmd.Lifecycle = "cnb"
				executeErr = cmd.Execute(nil)
			})

			It("prints warnings and helpful error message (that includes both names)", func() {
				Expect(testUI.Err).To(Say("a-warning"))
				Expect(testUI.Err).To(Say(`Buildpack 'the-buildpack' with stack 'stack!' with lifecycle 'cnb' does not exist\.`))
			})

		})
		When("deleting without a stack or lifecycle", func() {
			BeforeEach(func() {
				fakeActor.DeleteBuildpackByNameAndStackAndLifecycleReturns(v7action.Warnings{"a-warning"}, actionerror.BuildpackNotFoundError{BuildpackName: buildpackName})
				cmd.Stack = ""
				cmd.Lifecycle = ""
				executeErr = cmd.Execute(nil)
			})

			It("prints warnings and helpful error message", func() {
				Expect(testUI.Err).To(Say("a-warning"))
				Expect(testUI.Err).To(Say(`Buildpack 'the-buildpack' not found\.`))
			})
		})
	})

	It("delegates to the actor", func() {
		cmd.Stack = "the-stack"
		cmd.Lifecycle = "cnb"
		fakeActor.DeleteBuildpackByNameAndStackAndLifecycleReturns(nil, nil)

		executeErr = cmd.Execute(nil)

		Expect(executeErr).ToNot(HaveOccurred())
		actualBuildpack, actualStack, actualLifecycle := fakeActor.DeleteBuildpackByNameAndStackAndLifecycleArgsForCall(0)
		Expect(actualBuildpack).To(Equal("the-buildpack"))
		Expect(actualStack).To(Equal("the-stack"))
		Expect(actualLifecycle).To(Equal("cnb"))
	})

	It("prints warnings", func() {
		cmd.Stack = "a-stack"
		fakeActor.DeleteBuildpackByNameAndStackAndLifecycleReturns(v7action.Warnings{"a-warning"}, nil)

		executeErr = cmd.Execute(nil)

		Expect(executeErr).ToNot(HaveOccurred())
		Expect(testUI.Err).To(Say("a-warning"))
	})

	It("returns error from the actor and prints the errors", func() {
		cmd.Stack = "a-stack"

		fakeActor.DeleteBuildpackByNameAndStackAndLifecycleReturns(v7action.Warnings{"a-warning"}, errors.New("some-error"))

		executeErr = cmd.Execute(nil)

		Expect(executeErr).To(MatchError("some-error"))
		Expect(testUI.Err).To(Say("a-warning"))
	})
})
