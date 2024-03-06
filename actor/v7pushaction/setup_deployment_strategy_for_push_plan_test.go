package v7pushaction_test

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"

	. "code.cloudfoundry.org/cli/actor/v7pushaction"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SetupDeploymentStrategyForPushPlan", func() {
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
		expectedPushPlan, executeErr = SetupDeploymentStrategyForPushPlan(pushPlan, overrides)
	})

	When("flag overrides specifies strategy", func() {
		BeforeEach(func() {
			overrides.Strategy = "rolling"
		})

		It("sets the strategy on the push plan", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(expectedPushPlan.Strategy).To(Equal(constant.DeploymentStrategyRolling))
		})
	})

	When("flag overrides does not specify strategy", func() {
		It("leaves the strategy as its default value on the push plan", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(expectedPushPlan.Strategy).To(Equal(constant.DeploymentStrategyDefault))
		})
	})
})
