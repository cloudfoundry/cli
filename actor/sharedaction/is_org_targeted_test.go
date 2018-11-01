package sharedaction_test

import (
	. "code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/sharedaction/sharedactionfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("IsOrgTargeted", func() {
	var (
		actor      *Actor
		fakeConfig *sharedactionfakes.FakeConfig
	)

	BeforeEach(func() {
		fakeConfig = new(sharedactionfakes.FakeConfig)
		actor = NewActor(fakeConfig)
	})

	When("the config has an org targeted", func() {
		BeforeEach(func() {
			fakeConfig.HasTargetedOrganizationReturns(true)
		})

		It("returns true", func() {
			Expect(actor.IsOrgTargeted()).To(BeTrue())
		})
	})

	When("the config does not have an org targeted", func() {
		BeforeEach(func() {
			fakeConfig.HasTargetedOrganizationReturns(false)
		})

		It("returns false", func() {
			Expect(actor.IsOrgTargeted()).To(BeFalse())
		})
	})
})
