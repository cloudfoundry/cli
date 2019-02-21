package v7action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("FeatureFlag", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		actor, fakeCloudControllerClient, _, _, _ = NewTestActor()
	})

	Describe("GetFeatureFlagByName", func() {
		var (
			featureFlagName = "flag1"
			featureFlag     FeatureFlag
			warnings        Warnings
			executeErr      error
		)

		JustBeforeEach(func() {
			featureFlag, warnings, executeErr = actor.GetFeatureFlagByName(featureFlagName)
		})

		When("getting feature flag fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetFeatureFlagReturns(
					ccv3.FeatureFlag{},
					ccv3.Warnings{"this is a warning"},
					errors.New("some-error"))
			})

			It("returns warnings and error", func() {
				Expect(executeErr).To(MatchError("some-error"))
				Expect(warnings).To(ConsistOf("this is a warning"))
				Expect(fakeCloudControllerClient.GetFeatureFlagCallCount()).To(Equal(1))
				nameArg := fakeCloudControllerClient.GetFeatureFlagArgsForCall(0)
				Expect(nameArg).To(Equal(featureFlagName))
			})
		})

		When("no feature flag is returned", func() {
			BeforeEach(func() {
				var ccFeatureFlag ccv3.FeatureFlag

				fakeCloudControllerClient.GetFeatureFlagReturns(
					ccFeatureFlag,
					ccv3.Warnings{"this is a warning"},
					ccerror.FeatureFlagNotFoundError{})
			})

			It("returns warnings and a FeatureFlagNotFoundError", func() {
				Expect(executeErr).To(MatchError(actionerror.FeatureFlagNotFoundError{FeatureFlagName: featureFlagName}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})

		When("getting feature flag is successful", func() {

			BeforeEach(func() {
				featureFlagName = "flag1"
				ccFeatureFlag := ccv3.FeatureFlag{Name: "flag1"}
				fakeCloudControllerClient.GetFeatureFlagReturns(
					ccFeatureFlag,
					ccv3.Warnings{"this is a warning"},
					nil)
			})

			It("Returns the proper featureFlag", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))
				Expect(featureFlag).To(Equal(FeatureFlag{Name: "flag1"}))
			})
		})
	})

	Describe("GetFeatureFlags", func() {
		var (
			featureFlags []FeatureFlag
			warnings     Warnings
			executeErr   error
		)

		JustBeforeEach(func() {
			featureFlags, warnings, executeErr = actor.GetFeatureFlags()
		})

		When("The client is successful", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetFeatureFlagsReturns(
					[]ccv3.FeatureFlag{
						{
							Name:    "flag1",
							Enabled: false,
						},
						{
							Name:    "flag2",
							Enabled: true,
						},
					},
					ccv3.Warnings{"some-cc-warning"},
					nil,
				)
			})

			It("Returns the list of feature flags", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf(Warnings{"some-cc-warning"}))
				Expect(featureFlags).To(ConsistOf(
					FeatureFlag{
						Name:    "flag1",
						Enabled: false,
					},
					FeatureFlag{
						Name:    "flag2",
						Enabled: true,
					},
				))
			})
		})

		When("The client errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetFeatureFlagsReturns(nil, ccv3.Warnings{"some-cc-warning"}, errors.New("some-error"))
			})

			It("Returns the error", func() {
				Expect(executeErr).To(MatchError(errors.New("some-error")))
				Expect(warnings).To(ConsistOf("some-cc-warning"))
			})
		})
	})
})
