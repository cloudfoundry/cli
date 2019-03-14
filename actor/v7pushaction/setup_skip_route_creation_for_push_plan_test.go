package v7pushaction_test

import (
	"code.cloudfoundry.org/cli/util/manifestparser"

	. "code.cloudfoundry.org/cli/actor/v7pushaction"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SetupSkipRouteCreationForPushPlan", func() {
	var (
		pushPlan    PushPlan
		overrides   FlagOverrides
		manifestApp manifestparser.Application

		expectedPushPlan PushPlan
		executeErr       error
	)

	BeforeEach(func() {
		pushPlan = PushPlan{}
		overrides = FlagOverrides{}
		manifestApp = manifestparser.Application{}
	})

	JustBeforeEach(func() {
		expectedPushPlan, executeErr = SetupSkipRouteCreationForPushPlan(pushPlan, overrides, manifestApp)
	})

	When("flag overrides specifies skipping route creation", func() {
		BeforeEach(func() {
			overrides.SkipRouteCreation = true
		})

		It("sets SkipRouteCreation on the push plan", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(expectedPushPlan.SkipRouteCreation).To(BeTrue())
		})
	})
})
