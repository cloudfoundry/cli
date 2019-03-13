package v7pushaction_test

import (
	"code.cloudfoundry.org/cli/util/manifestparser"

	. "code.cloudfoundry.org/cli/actor/v7pushaction"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SetupNoStartForPushPlan", func() {
	var (
		pushPlan    PushPlan
		manifestApp manifestparser.Application

		expectedPushPlan PushPlan
		executeErr       error
	)

	BeforeEach(func() {
		pushPlan = PushPlan{}
		manifestApp = manifestparser.Application{}
	})

	JustBeforeEach(func() {
		expectedPushPlan, executeErr = SetupNoStartForPushPlan(pushPlan, manifestApp)
	})

	When("flag overrides specifies no start", func() {
		BeforeEach(func() {
			pushPlan.Overrides.NoStart = true
		})

		It("sets no start on the push plan", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(expectedPushPlan.NoStart).To(BeTrue())
		})
	})
})
