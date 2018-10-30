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
		binaryName string
		fakeConfig *sharedactionfakes.FakeConfig
	)

	BeforeEach(func() {
		binaryName = "faceman"
		fakeConfig = new(sharedactionfakes.FakeConfig)
		fakeConfig.BinaryNameReturns(binaryName)
		actor = NewActor(fakeConfig)
	})

	When("the config has an org targeted", func() {
		It("returns true", func() {
			fakeConfig.HasTargetedOrganizationReturns(true)

			Expect(actor.IsOrgTargeted()).To(BeTrue())
		})
	})

	When("the config does not have an org targeted", func() {
		It("returns false", func() {
			fakeConfig.HasTargetedOrganizationReturns(false)

			Expect(actor.IsOrgTargeted()).To(BeFalse())
		})
	})
})
