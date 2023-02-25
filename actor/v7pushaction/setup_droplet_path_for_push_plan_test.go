package v7pushaction_test

import (
	. "code.cloudfoundry.org/cli/actor/v7pushaction"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SetupDropletPathForPushPlan", func() {
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
		expectedPushPlan, executeErr = SetupDropletPathForPushPlan(pushPlan, overrides)
	})

	When("flag overrides specifies droplet path", func() {
		BeforeEach(func() {
			overrides.DropletPath = "some-droplet.tgz"
		})

		It("sets the droplet path on the push plan", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(expectedPushPlan.DropletPath).To(Equal("some-droplet.tgz"))
		})
	})

	When("flag overrides does not specify droplet path", func() {
		It("leaves the droplet path as its default value on the push plan", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(expectedPushPlan.DropletPath).To(Equal(""))
		})
	})
})
