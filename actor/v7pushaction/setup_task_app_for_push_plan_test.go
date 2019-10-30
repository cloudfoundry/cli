package v7pushaction_test

import (
	. "code.cloudfoundry.org/cli/actor/v7pushaction"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SetupTaskAppForPushPlan", func() {
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
		expectedPushPlan, executeErr = SetupTaskAppForPushPlan(pushPlan, overrides)
	})

	When("flag overrides specifies task type app", func() {
		BeforeEach(func() {
			overrides.Task = true
		})

		It("sets task app type on the push plan", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(expectedPushPlan.TaskTypeApplication).To(BeTrue())
		})
	})
})
