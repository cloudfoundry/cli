package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/v8/actor/v7action"
	"code.cloudfoundry.org/cli/v8/resources"

	"code.cloudfoundry.org/cli/v8/actor/actionerror"
	"code.cloudfoundry.org/cli/v8/command/commandfakes"
	. "code.cloudfoundry.org/cli/v8/command/v7"
	"code.cloudfoundry.org/cli/v8/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/v8/util/configv3"
	"code.cloudfoundry.org/cli/v8/util/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("Stack Command", func() {
	var (
		cmd             StackCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		binaryName      string
		executeErr      error
		stackName       string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		cmd = StackCommand{
			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		stackName = "some-stack-name"

		cmd.RequiredArgs.StackName = stackName
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	Context("When the environment is not setup correctly", func() {
		When("checking target fails", func() {
			BeforeEach(func() {
				fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: binaryName})
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: binaryName}))

				Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
				checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
				Expect(fakeActor.GetStackByNameCallCount()).To(Equal(0))
				Expect(checkTargetedOrg).To(BeFalse())
				Expect(checkTargetedSpace).To(BeFalse())
			})
		})

		When("retrieving user information errors", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("some current user error")
				fakeActor.GetCurrentUserReturns(configv3.User{}, expectedErr)
			})

			It("return an error", func() {
				Expect(executeErr).To(Equal(expectedErr))
			})
		})
	})

	Context("When the environment is setup correctly", func() {
		BeforeEach(func() {
			fakeConfig.TargetedSpaceReturns(configv3.Space{Name: "some-space", GUID: "some-space-guid"})
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "some-org"})
			fakeActor.GetCurrentUserReturns(configv3.User{Name: "banana"}, nil)
		})

		Context("When the stack exists", func() {
			BeforeEach(func() {
				stack := resources.Stack{
					Name:        "some-stack-name",
					GUID:        "some-stack-guid",
					Description: "some-stack-desc",
				}
				fakeActor.GetStackByNameReturns(stack, v7action.Warnings{"some-warning-1"}, nil)
			})

			When("The --guid flag is not provided", func() {
				It("Displays the stack information", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(fakeActor.GetStackByNameArgsForCall(0)).To(Equal("some-stack-name"))
					Expect(fakeActor.GetStackByNameCallCount()).To(Equal(1))
					// NOTE: DISPLAY EXPECTS
					Expect(testUI.Err).To(Say("some-warning-1"))
				})
			})

			When("The --guid flag is provided", func() {
				BeforeEach(func() {
					cmd.GUID = true
				})

				It("displays just the guid", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(fakeActor.GetStackByNameArgsForCall(0)).To(Equal("some-stack-name"))
					Expect(fakeActor.GetStackByNameCallCount()).To(Equal(1))
					Expect(testUI.Err).To(Say("some-warning-1"))
				})
			})
		})

		Context("When the stack has no state", func() {
			BeforeEach(func() {
				stack := resources.Stack{
					Name:        "some-stack-name",
					GUID:        "some-stack-guid",
					Description: "some-stack-desc",
				}
				fakeActor.GetStackByNameReturns(stack, v7action.Warnings{}, nil)
			})

			It("Displays the stack information without state", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeActor.GetStackByNameArgsForCall(0)).To(Equal("some-stack-name"))
				Expect(fakeActor.GetStackByNameCallCount()).To(Equal(1))

				Expect(testUI.Out).To(Say("name:\\s+some-stack-name"))
				Expect(testUI.Out).To(Say("description:\\s+some-stack-desc"))
				Expect(testUI.Out).NotTo(Say("state:"))
			})
		})

	Context("When the stack has a state", func() {
		Context("When the state is ACTIVE", func() {
			BeforeEach(func() {
				stack := resources.Stack{
					Name:        "some-stack-name",
					GUID:        "some-stack-guid",
					Description: "some-stack-desc",
					State:       "ACTIVE",
				}
				fakeActor.GetStackByNameReturns(stack, v7action.Warnings{}, nil)
			})

			It("Displays the stack information with state but no reason", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeActor.GetStackByNameArgsForCall(0)).To(Equal("some-stack-name"))
				Expect(fakeActor.GetStackByNameCallCount()).To(Equal(1))

				Expect(testUI.Out).To(Say("name:\\s+some-stack-name"))
				Expect(testUI.Out).To(Say("description:\\s+some-stack-desc"))
				Expect(testUI.Out).To(Say("state:\\s+ACTIVE"))
				Expect(testUI.Out).NotTo(Say("reason:"))
			})
		})

		Context("When the state is not ACTIVE and has a reason", func() {
			BeforeEach(func() {
				stack := resources.Stack{
					Name:        "some-stack-name",
					GUID:        "some-stack-guid",
					Description: "some-stack-desc",
					State:       "DEPRECATED",
					StateReason: "This stack is being phased out",
				}
				fakeActor.GetStackByNameReturns(stack, v7action.Warnings{}, nil)
			})

			It("Displays the stack information with state and reason", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeActor.GetStackByNameArgsForCall(0)).To(Equal("some-stack-name"))
				Expect(fakeActor.GetStackByNameCallCount()).To(Equal(1))

				Expect(testUI.Out).To(Say("name:\\s+some-stack-name"))
				Expect(testUI.Out).To(Say("description:\\s+some-stack-desc"))
				Expect(testUI.Out).To(Say("state:\\s+DEPRECATED"))
				Expect(testUI.Out).To(Say("reason:\\s+This stack is being phased out"))
			})
		})

		Context("When the state is not ACTIVE but has no reason", func() {
			BeforeEach(func() {
				stack := resources.Stack{
					Name:        "some-stack-name",
					GUID:        "some-stack-guid",
					Description: "some-stack-desc",
					State:       "RESTRICTED",
				}
				fakeActor.GetStackByNameReturns(stack, v7action.Warnings{}, nil)
			})

			It("Displays the stack information with state and empty reason", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeActor.GetStackByNameArgsForCall(0)).To(Equal("some-stack-name"))
				Expect(fakeActor.GetStackByNameCallCount()).To(Equal(1))

				Expect(testUI.Out).To(Say("name:\\s+some-stack-name"))
				Expect(testUI.Out).To(Say("description:\\s+some-stack-desc"))
				Expect(testUI.Out).To(Say("state:\\s+RESTRICTED"))
				Expect(testUI.Out).To(Say("reason:"))
			})
		})
	})

		When("The Stack does not Exist", func() {
			expectedError := actionerror.StackNotFoundError{Name: "some-stack-name"}
			BeforeEach(func() {
				fakeActor.GetStackByNameReturns(
					resources.Stack{},
					v7action.Warnings{"some-warning-1"},
					expectedError,
				)
			})

			It("Fails and returns a StackNotFoundError", func() {
				Expect(fakeActor.GetStackByNameArgsForCall(0)).To(Equal("some-stack-name"))
				Expect(fakeActor.GetStackByNameCallCount()).To(Equal(1))
				Expect(executeErr).To(Equal(expectedError))
				Expect(testUI.Err).To(Say("some-warning-1"))
			})
		})

		When("There was an error in the actor", func() {
			BeforeEach(func() {
				fakeActor.GetStackByNameReturns(
					resources.Stack{},
					v7action.Warnings{"some-warning-1"},
					errors.New("some-random-error"),
				)
			})

			It("Fails and returns a StackNotFoundError", func() {
				Expect(fakeActor.GetStackByNameArgsForCall(0)).To(Equal("some-stack-name"))
				Expect(fakeActor.GetStackByNameCallCount()).To(Equal(1))
				Expect(executeErr).To(MatchError(errors.New("some-random-error")))
				Expect(testUI.Err).To(Say("some-warning-1"))
			})
		})

	})

})
