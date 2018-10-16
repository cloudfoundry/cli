package sharedaction_test

import (
	. "code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/sharedaction/sharedactionfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CheckOrgSpaceTargeted", func() {
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

	When("the user is not logged in", func() {
		It("returns false", func() {
			targeted := actor.CheckOrgSpaceTargeted()
			Expect(targeted).To(BeFalse())
		})
	})

	When("the user is logged in", func() {
		BeforeEach(func() {
			fakeConfig.AccessTokenReturns("some-access-token")
			fakeConfig.RefreshTokenReturns("some-refresh-token")
		})

		Context("and both an org and space are targeted", func() {
			It("returns true", func() {
				fakeConfig.HasTargetedOrganizationReturns(true)
				fakeConfig.HasTargetedSpaceReturns(true)

				targeted := actor.CheckOrgSpaceTargeted()
				Expect(targeted).To(BeTrue())
			})
		})

		Context("only org is targeted", func() {
			It("returns false", func() {
				fakeConfig.HasTargetedOrganizationReturns(true)
				fakeConfig.HasTargetedSpaceReturns(false)

				targeted := actor.CheckOrgSpaceTargeted()
				Expect(targeted).To(BeFalse())
			})
		})

		Context("neither org or space is targeted", func() {
			It("returns false", func() {
				fakeConfig.HasTargetedOrganizationReturns(false)
				fakeConfig.HasTargetedSpaceReturns(false)

				targeted := actor.CheckOrgSpaceTargeted()
				Expect(targeted).To(BeFalse())
			})
		})
	})
})
