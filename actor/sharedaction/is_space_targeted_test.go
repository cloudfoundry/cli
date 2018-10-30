package sharedaction_test

import (
	. "code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/sharedaction/sharedactionfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("IsSpaceTargeted", func() {
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

	When("the config has a space targeted", func() {
		It("returns true", func() {
			fakeConfig.HasTargetedSpaceReturns(true)

			Expect(actor.IsSpaceTargeted()).To(BeTrue())
		})
	})

	When("the config does not have a space targeted", func() {
		It("returns false", func() {
			fakeConfig.HasTargetedSpaceReturns(false)

			Expect(actor.IsSpaceTargeted()).To(BeFalse())
		})
	})
})
