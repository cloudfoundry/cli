package sharedaction_test

import (
	. "code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/sharedaction/sharedactionfakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("IsSpaceTargeted", func() {
	var (
		actor      *Actor
		fakeConfig *sharedactionfakes.FakeConfig
	)

	BeforeEach(func() {
		fakeConfig = new(sharedactionfakes.FakeConfig)
		actor = NewActor(fakeConfig)
	})

	When("the config has a space targeted", func() {
		BeforeEach(func() {
			fakeConfig.HasTargetedSpaceReturns(true)
		})

		It("returns true", func() {
			Expect(actor.IsSpaceTargeted()).To(BeTrue())
		})
	})

	When("the config does not have a space targeted", func() {
		BeforeEach(func() {
			fakeConfig.HasTargetedSpaceReturns(false)
		})

		It("returns false", func() {
			Expect(actor.IsSpaceTargeted()).To(BeFalse())
		})
	})
})
