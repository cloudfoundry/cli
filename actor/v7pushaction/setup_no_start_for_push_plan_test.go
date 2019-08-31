package v7pushaction_test

import (
	. "code.cloudfoundry.org/cli/actor/v7pushaction"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SetupNoStartForPushPlan", func() {
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
		expectedPushPlan, executeErr = SetupNoStartForPushPlan(pushPlan, overrides)
	})

	When("flag overrides specifies no start", func() {
		BeforeEach(func() {
			overrides.NoStart = true
		})

		It("sets no start on the push plan", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(expectedPushPlan.NoStart).To(BeTrue())
		})
	})
})
