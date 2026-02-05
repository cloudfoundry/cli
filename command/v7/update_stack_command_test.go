package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/v9/actor/actionerror"
	"code.cloudfoundry.org/cli/v9/actor/v7action"
	"code.cloudfoundry.org/cli/v9/command/commandfakes"
	. "code.cloudfoundry.org/cli/v9/command/v7"
	"code.cloudfoundry.org/cli/v9/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/v9/resources"
	"code.cloudfoundry.org/cli/v9/util/configv3"
	"code.cloudfoundry.org/cli/v9/util/ui"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("update-stack Command", func() {
	var (
		cmd             UpdateStackCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		executeErr      error
		args            []string
		binaryName      string
	)

	JustBeforeEach(func() {
		executeErr = cmd.Execute(args)
	})

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)
		args = nil

		cmd = UpdateStackCommand{
			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		fakeConfig.APIVersionReturns("3.210.0")
	})

	Context("When the environment is not setup correctly", func() {
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

	Context("When the environment is setup correctly", func() {
		BeforeEach(func() {
			fakeActor.GetCurrentUserReturns(configv3.User{Name: "banana"}, nil)
			cmd.RequiredArgs.StackName = "some-stack"
		})

		Context("when the state flag is invalid", func() {
			BeforeEach(func() {
				cmd.State = "invalid"
			})

			It("returns an error", func() {
				Expect(executeErr).To(HaveOccurred())
				Expect(executeErr.Error()).To(ContainSubstring("Invalid state"))
			})
		})

		Context("when the state flag is valid", func() {
			BeforeEach(func() {
				cmd.State = "deprecated"
			})

			Context("when getting the stack fails", func() {
				BeforeEach(func() {
					fakeActor.GetStackByNameReturns(resources.Stack{}, v7action.Warnings{"warning-1"}, errors.New("get-stack-error"))
				})

				It("returns the error and displays warnings", func() {
					Expect(executeErr).To(MatchError("get-stack-error"))
					Expect(testUI.Err).To(Say("warning-1"))
				})
			})

			Context("when getting the stack succeeds", func() {
				BeforeEach(func() {
					fakeActor.GetStackByNameReturns(resources.Stack{
						GUID: "stack-guid",
						Name: "some-stack",
					}, v7action.Warnings{"warning-1"}, nil)
				})

				Context("when updating the stack fails", func() {
					BeforeEach(func() {
						fakeActor.UpdateStackReturns(resources.Stack{}, v7action.Warnings{"warning-2"}, errors.New("update-error"))
					})

					It("returns the error and displays warnings", func() {
						Expect(executeErr).To(MatchError("update-error"))
						Expect(testUI.Err).To(Say("warning-1"))
						Expect(testUI.Err).To(Say("warning-2"))
						Expect(testUI.Out).To(Say("Updating stack some-stack as banana..."))
					})
				})

				Context("when updating the stack succeeds", func() {
					BeforeEach(func() {
						fakeActor.UpdateStackReturns(resources.Stack{
							GUID:        "stack-guid",
							Name:        "some-stack",
							Description: "some description",
							State:       resources.StackStateDeprecated,
						}, v7action.Warnings{"warning-2"}, nil)
					})

					It("displays the updated stack info", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say("Updating stack some-stack as banana..."))
						Expect(testUI.Out).To(Say("OK"))
						Expect(testUI.Out).To(Say(`name:\s+some-stack`))
						Expect(testUI.Out).To(Say(`description:\s+some description`))
						Expect(testUI.Out).To(Say(`state:\s+DEPRECATED`))

						Expect(testUI.Err).To(Say("warning-1"))
						Expect(testUI.Err).To(Say("warning-2"))

						Expect(fakeActor.GetStackByNameCallCount()).To(Equal(1))
						Expect(fakeActor.GetStackByNameArgsForCall(0)).To(Equal("some-stack"))

						Expect(fakeActor.UpdateStackCallCount()).To(Equal(1))
						guid, state, reason := fakeActor.UpdateStackArgsForCall(0)
						Expect(guid).To(Equal("stack-guid"))
						Expect(state).To(Equal(resources.StackStateDeprecated))
						Expect(reason).To(Equal(""))
					})
				})
			})
		})

	Context("when state values are provided in different cases", func() {
		It("accepts 'active' and capitalizes it", func() {
			cmd.State = "active"
			fakeActor.GetStackByNameReturns(resources.Stack{GUID: "guid"}, v7action.Warnings{}, nil)
			fakeActor.UpdateStackReturns(resources.Stack{Name: "some-stack", State: resources.StackStateActive}, v7action.Warnings{}, nil)

			executeErr = cmd.Execute(args)

			Expect(executeErr).ToNot(HaveOccurred())
			_, state, _ := fakeActor.UpdateStackArgsForCall(0)
			Expect(state).To(Equal(resources.StackStateActive))
		})

		It("accepts 'RESTRICTED' and keeps it capitalized", func() {
			cmd.State = "RESTRICTED"
			fakeActor.GetStackByNameReturns(resources.Stack{GUID: "guid"}, v7action.Warnings{}, nil)
			fakeActor.UpdateStackReturns(resources.Stack{Name: "some-stack", State: resources.StackStateRestricted}, v7action.Warnings{}, nil)

			executeErr = cmd.Execute(args)

			Expect(executeErr).ToNot(HaveOccurred())
			_, state, _ := fakeActor.UpdateStackArgsForCall(0)
			Expect(state).To(Equal(resources.StackStateRestricted))
		})

		It("accepts 'Disabled' and capitalizes it", func() {
			cmd.State = "Disabled"
			fakeActor.GetStackByNameReturns(resources.Stack{GUID: "guid"}, v7action.Warnings{}, nil)
			fakeActor.UpdateStackReturns(resources.Stack{Name: "some-stack", State: resources.StackStateDisabled}, v7action.Warnings{}, nil)

			executeErr = cmd.Execute(args)

			Expect(executeErr).ToNot(HaveOccurred())
			_, state, _ := fakeActor.UpdateStackArgsForCall(0)
			Expect(state).To(Equal(resources.StackStateDisabled))
		})
	})

	Context("when the reason flag is provided", func() {
		BeforeEach(func() {
			cmd.State = "deprecated"
			cmd.Reason = "Use cflinuxfs4 instead"
			fakeActor.GetStackByNameReturns(resources.Stack{GUID: "guid"}, v7action.Warnings{}, nil)
			fakeActor.UpdateStackReturns(resources.Stack{
				Name:        "some-stack",
				Description: "some description",
				State:       resources.StackStateDeprecated,
				StateReason: "Use cflinuxfs4 instead",
			}, v7action.Warnings{}, nil)
		})

		It("passes the reason to the actor and displays it", func() {
			executeErr = cmd.Execute(args)

			Expect(executeErr).ToNot(HaveOccurred())

			Expect(fakeActor.UpdateStackCallCount()).To(Equal(1))
			_, _, reason := fakeActor.UpdateStackArgsForCall(0)
			Expect(reason).To(Equal("Use cflinuxfs4 instead"))

			Expect(testUI.Out).To(Say(`reason:\s+Use cflinuxfs4 instead`))
		})
	})
	})
})

