package v7pushaction_test

import (
	. "code.cloudfoundry.org/cli/actor/v7pushaction"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("IsDropletPathSet", func() {
	var (
		pushPlan     PushPlan
		returnedBool bool
		actor        Actor
	)
	JustBeforeEach(func() {
		returnedBool = actor.IsDropletPathSet(pushPlan)
	})

	When("the droplet path is set", func() {
		BeforeEach(func() {
			pushPlan = PushPlan{
				DropletPath: "test",
			}
		})

		It("returns true", func() {
			Expect(returnedBool).To(BeTrue())
		})
	})

	When("the conditional is false", func() {
		BeforeEach(func() {
			pushPlan = PushPlan{
				DropletPath: "",
			}
		})

		It("returns false", func() {
			Expect(returnedBool).To(BeFalse())
		})
	})
})
