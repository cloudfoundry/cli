package v7pushaction_test

import (
	. "code.cloudfoundry.org/cli/actor/v7pushaction"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SetupNoWaitForPushPlan", func() {
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
		expectedPushPlan, executeErr = SetupNoWaitForPushPlan(pushPlan, overrides)
	})

	When("flag override specifies no-wait", func() {
		BeforeEach(func() {
			overrides.NoWait = true
		})

		It("sets the no-wait flag on the push plan", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(expectedPushPlan.NoWait).To(Equal(true))
		})
	})

	When("flag overrides does not specify no-wait", func() {
		It("leaves the no-wait flag as false on the push plan", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(expectedPushPlan.NoWait).To(Equal(false))
		})
	})
})
