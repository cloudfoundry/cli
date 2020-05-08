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

var _ = Describe("Feature Flags Command", func() {
	var (
		cmd             FeatureFlagsCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		executeErr      error
		binaryName      string
	)

	const tableHeaders = `name\s+state`

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		cmd = FeatureFlagsCommand{
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

	Context("When the environment is not setup correctly", func() {
		When("checking target fails", func() {
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
	})

	Context("When the environment is setup correctly", func() {
		BeforeEach(func() {
			fakeConfig.CurrentUserReturns(configv3.User{Name: "banana"}, nil)
		})

		When("FeatureFlagsActor returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				warnings := v7action.Warnings{"warning-1", "warning-2"}
				expectedErr = errors.New("some-error")
				fakeActor.GetFeatureFlagsReturns(nil, warnings, expectedErr)
			})

			It("prints that error with warnings", func() {
				Expect(executeErr).To(Equal(expectedErr))

				Expect(testUI.Err).To(Say("warning-1"))
				Expect(testUI.Err).To(Say("warning-2"))
				Expect(testUI.Out).ToNot(Say(tableHeaders))
			})
		})

		When("everything is perfect", func() {
			BeforeEach(func() {
				flags := []v7action.FeatureFlag{
					{Name: "flag2", Enabled: true},
					{Name: "flag1", Enabled: false},
				}
				fakeActor.GetFeatureFlagsReturns(flags, v7action.Warnings{"warning-1", "warning-2"}, nil)
			})

			It("asks the FeatureFlagsActor for a list of feature flags", func() {
				Expect(fakeActor.GetFeatureFlagsCallCount()).To(Equal(1))
			})

			It("prints warnings", func() {
				Expect(testUI.Err).To(Say(`warning-1`))
				Expect(testUI.Err).To(Say(`warning-2`))
			})

			It("prints the list of feature flags in alphabetical order", func() {
				Expect(testUI.Out).To(Say(tableHeaders))
				Expect(testUI.Out).To(Say(`flag2\s+enabled`))
				Expect(testUI.Out).To(Say(`flag1\s+disabled`))
			})

			It("prints the flavor text", func() {
				Expect(testUI.Out).To(Say("Getting feature flags as banana\\.\\.\\."))
			})
		})
	})
})
