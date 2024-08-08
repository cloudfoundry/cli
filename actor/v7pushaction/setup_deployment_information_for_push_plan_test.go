package v7pushaction_test

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"

	. "code.cloudfoundry.org/cli/actor/v7pushaction"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SetupDeploymentInformationForPushPlan", func() {
	var (
		pushPlan  PushPlan
		overrides FlagOverrides

		expectedPushPlan PushPlan
		executeErr       error
	)

	BeforeEach(func() {
		pushPlan = PushPlan{}
		overrides = FlagOverrides{}
	})

	JustBeforeEach(func() {
		expectedPushPlan, executeErr = SetupDeploymentInformationForPushPlan(pushPlan, overrides)
	})

	When("flag overrides specifies strategy", func() {
		BeforeEach(func() {
			overrides.Strategy = "rolling"
			overrides.MaxInFlight = 5
		})

		It("sets the strategy on the push plan", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(expectedPushPlan.Strategy).To(Equal(constant.DeploymentStrategyRolling))
		})

		It("sets the max in flight on the push plan", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(expectedPushPlan.MaxInFlight).To(Equal(5))
		})
	})

	When("flag overrides does not specify strategy", func() {
		BeforeEach(func() {
			overrides.MaxInFlight = 10
		})
		It("leaves the strategy as its default value on the push plan", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(expectedPushPlan.Strategy).To(Equal(constant.DeploymentStrategyDefault))
		})

		It("does not set MaxInFlight", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(expectedPushPlan.MaxInFlight).To(Equal(0))
		})
	})
})
