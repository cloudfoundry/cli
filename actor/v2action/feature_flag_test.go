package v2action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Feature Flag Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v2actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil)
	})

	Describe("GetFeatureFlags", func() {
		var (
			featureFlags []FeatureFlag
			warnings     Warnings
			err          error
		)

		JustBeforeEach(func() {
			featureFlags, warnings, err = actor.GetFeatureFlags()
		})

		Context("when an error is encountered getting feature flags", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("get feature flags error")
				fakeCloudControllerClient.GetConfigFeatureFlagsReturns(
					[]ccv2.FeatureFlag{},
					ccv2.Warnings{"get-flags-warning"},
					expectedErr)
			})

			It("returns the error and all warnings", func() {
				Expect(err).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("get-flags-warning"))

				Expect(fakeCloudControllerClient.GetConfigFeatureFlagsCallCount()).To(Equal(1))
			})
		})

		Context("when no errors are encountered getting feature flags", func() {
			var featureFlag1 ccv2.FeatureFlag
			var featureFlag2 ccv2.FeatureFlag

			BeforeEach(func() {
				featureFlag1 = ccv2.FeatureFlag{Name: "feature-flag-1", Enabled: true}
				featureFlag2 = ccv2.FeatureFlag{Name: "feature-flag-2", Enabled: false}
				fakeCloudControllerClient.GetConfigFeatureFlagsReturns(
					[]ccv2.FeatureFlag{featureFlag1, featureFlag2},
					ccv2.Warnings{"get-flags-warning"},
					nil)
			})

			It("returns the feature flags and all warnings", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("get-flags-warning"))
				Expect(featureFlags).To(Equal([]FeatureFlag{
					FeatureFlag(featureFlag1),
					FeatureFlag(featureFlag2),
				}))

				Expect(fakeCloudControllerClient.GetConfigFeatureFlagsCallCount()).To(Equal(1))
			})
		})
	})
})
