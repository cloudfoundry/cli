package v2_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v2"
	"code.cloudfoundry.org/cli/command/v2/v2fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("feature flags Command", func() {
	var (
		cmd             FeatureFlagsCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v2fakes.FakeFeatureFlagsActor
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v2fakes.FakeFeatureFlagsActor)

		cmd = FeatureFlagsCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	Context("when the user is not logged in", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: binaryName}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrgArg, checkTargetedSpaceArg := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrgArg).To(BeFalse())
			Expect(checkTargetedSpaceArg).To(BeFalse())
		})
	})

	Context("when the user is logged in", func() {
		Context("when getting the current user fails", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserReturns(configv3.User{}, errors.New("get-user-error"))
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError("get-user-error"))
			})
		})

		Context("when getting the current user succeeds", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserReturns(
					configv3.User{Name: "some-user"},
					nil)
			})

			Context("when an error is encountered getting feature flags", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("get feature flags error")
					fakeActor.GetFeatureFlagsReturns(
						[]v2action.FeatureFlag{},
						v2action.Warnings{"get-flags-warning"},
						expectedErr)
				})

				It("displays an empty list and all warnings", func() {
					Expect(executeErr).To(MatchError(expectedErr))

					Expect(testUI.Err).To(Say("get-flags-warning"))

					Expect(fakeActor.GetFeatureFlagsCallCount()).To(Equal(1))
				})
			})

			Context("when there are no feature flags", func() {
				BeforeEach(func() {
					fakeActor.GetFeatureFlagsReturns(
						[]v2action.FeatureFlag{},
						v2action.Warnings{"get-flags-warning"},
						nil)
				})

				It("displays an empty list and all warnings", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say("Retrieving status of all flagged features as some-user\\.\\.\\."))
					Expect(testUI.Out).To(Say(""))
					Expect(testUI.Out).To(Say("features\\s+state"))

					Expect(testUI.Err).To(Say("get-flags-warning"))

					Expect(fakeActor.GetFeatureFlagsCallCount()).To(Equal(1))
				})
			})

			Context("when there are feature flags", func() {
				BeforeEach(func() {
					fakeActor.GetFeatureFlagsReturns(
						[]v2action.FeatureFlag{
							{
								Name:    "feature-flag-1",
								Enabled: true,
							},
							{
								Name:    "feature-flag-2",
								Enabled: false,
							},
						},
						v2action.Warnings{"get-flags-warning"},
						nil)
				})

				It("displays a list of feature flags with state and all warnings", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say("Retrieving status of all flagged features as some-user\\.\\.\\."))
					Expect(testUI.Out).To(Say(""))
					Expect(testUI.Out).To(Say("features\\s+state"))
					Expect(testUI.Out).To(Say("feature-flag-1\\s+enabled"))
					Expect(testUI.Out).To(Say("feature-flag-2\\s+disabled"))

					Expect(testUI.Err).To(Say("get-flags-warning"))

					Expect(fakeActor.GetFeatureFlagsCallCount()).To(Equal(1))
				})
			})
		})
	})
})
