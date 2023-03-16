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

var _ = Describe("Disable Feature Flag Command", func() {
	var (
		cmd             DisableFeatureFlagCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		executeErr      error
		binaryName      string
		featureFlagName = "flag1"
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		cmd = DisableFeatureFlagCommand{
			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)

		cmd.RequiredArgs.Feature = featureFlagName
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
			Expect(testUI.Out).To(Say(`Disabling feature flag flag1 as apple\.\.\.`))
		})

		When("updating featureFlag fails", func() {
			BeforeEach(func() {
				fakeActor.DisableFeatureFlagReturns(v7action.Warnings{"this is a warning"},
					errors.New("some-error"))
			})

			It("prints warnings and returns error", func() {
				Expect(executeErr).To(MatchError("some-error"))
				Expect(testUI.Err).To(Say("this is a warning"))
			})
		})

		When("updating featureFlag succeeds", func() {
			When("featureFlag exist", func() {
				BeforeEach(func() {
					fakeActor.DisableFeatureFlagReturns(v7action.Warnings{"this is a warning"}, nil)
				})

				It("displays the feature flag was enabled", func() {
					featureFlagArgs := fakeActor.DisableFeatureFlagArgsForCall(0)
					Expect(featureFlagArgs).To(Equal(featureFlagName))
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(testUI.Err).To(Say("this is a warning"))
					Expect(testUI.Out).To(Say(`OK`))
				})
			})
			When("there is no featureFlag", func() {
				BeforeEach(func() {
					fakeActor.DisableFeatureFlagReturns(v7action.Warnings{"this is a warning"}, actionerror.FeatureFlagNotFoundError{FeatureFlagName: featureFlagName})
				})

				It("Fails and returns a FeatureFlagNotFoundError", func() {
					Expect(executeErr).To(Equal(actionerror.FeatureFlagNotFoundError{FeatureFlagName: featureFlagName}))
					Expect(testUI.Err).To(Say("this is a warning"))
				})
			})
		})
	})
})
