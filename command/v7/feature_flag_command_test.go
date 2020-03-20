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

var _ = Describe("Feature Flag Command", func() {
	var (
		cmd             FeatureFlagCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeFeatureFlagActor
		executeErr      error
		args            []string
		binaryName      string
		featureFlagName = "flag1"
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeFeatureFlagActor)
		args = nil

		cmd = FeatureFlagCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)

		cmd.RequiredArgs.Feature = featureFlagName
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
			Expect(testUI.Out).To(Say(`Getting info for feature flag flag1 as apple\.\.\.`))
		})

		When("getting featureFlag fails", func() {
			BeforeEach(func() {
				fakeActor.GetFeatureFlagByNameReturns(v7action.FeatureFlag{}, v7action.Warnings{"this is a warning"},
					errors.New("some-error"))
			})

			It("prints warnings and returns error", func() {
				Expect(executeErr).To(MatchError("some-error"))
				Expect(testUI.Err).To(Say("this is a warning"))
			})
		})

		When("getting featureFlag succeeds", func() {
			When("featureFlag exist", func() {
				BeforeEach(func() {
					featureFlag := v7action.FeatureFlag{
						Name:    "flag1",
						Enabled: true,
					}
					fakeActor.GetFeatureFlagByNameReturns(featureFlag, v7action.Warnings{"this is a warning"}, nil)
				})

				It("prints a table of featureFlag", func() {
					featureFlagArgs := fakeActor.GetFeatureFlagByNameArgsForCall(0)
					Expect(featureFlagArgs).To(Equal(featureFlagName))
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(testUI.Err).To(Say("this is a warning"))
					Expect(testUI.Out).To(Say(`Features\s+State`))
					Expect(testUI.Out).To(Say(`flag1\s+enabled`))
				})
			})
			When("there is no featureFlag", func() {
				BeforeEach(func() {
					featureFlag := v7action.FeatureFlag{}
					fakeActor.GetFeatureFlagByNameReturns(featureFlag, v7action.Warnings{"this is a warning"}, actionerror.FeatureFlagNotFoundError{FeatureFlagName: featureFlagName})
				})

				It("Fails and returns a FeatureFlagNotFoundError", func() {
					Expect(executeErr).To(Equal(actionerror.FeatureFlagNotFoundError{FeatureFlagName: featureFlagName}))
					Expect(testUI.Err).To(Say("this is a warning"))
				})
			})
		})
	})
})
