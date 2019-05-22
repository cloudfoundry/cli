package v7pushaction_test

import (
	. "code.cloudfoundry.org/cli/actor/v7pushaction"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ConditionallyRunFunc", func() {
	var (
		conditional      func(plan PushPlan) bool
		pushPlan         PushPlan
		returnedPushPlan PushPlan
		changeFunc       ChangeApplicationFunc
		actor            Actor
		eventStream      chan<- Event
		progressBar      ProgressBar
	)
	JustBeforeEach(func() {
		var err error
		returnedPushPlan, _, err = actor.ConditionallyRunFunc(conditional, changeFunc)(pushPlan, eventStream, progressBar)
		Expect(err).NotTo(HaveOccurred())
	})

	BeforeEach(func() {
		conditional = func(plan PushPlan) bool {
			return plan.DropletPath == "test"
		}

		changeFunc = func(pushPlan PushPlan, eventStream chan<- Event, progressBar ProgressBar) (plan PushPlan, warnings Warnings, e error) {
			pushPlan.DropletGUID = "updated"
			return pushPlan, nil, nil
		}
	})

	When("the conditional is true", func() {
		BeforeEach(func() {
			pushPlan = PushPlan{
				DropletPath: "test",
			}
		})

		It("runs the change func", func() {
			Expect(returnedPushPlan.DropletGUID).To(Equal("updated"))
		})
	})

	When("the conditional is false", func() {
		BeforeEach(func() {
			pushPlan = PushPlan{
				DropletPath: "not test",
			}
		})

		It("does not run the change func", func() {
			Expect(returnedPushPlan.DropletGUID).NotTo(Equal("updated"))
		})
	})
})
