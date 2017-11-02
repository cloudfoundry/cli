package sharedaction_test

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/sharedaction/sharedactionfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("CheckTarget", func() {
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

	Context("when the user is not logged in", func() {
		It("returns an error", func() {
			err := actor.CheckTarget(false, false)
			Expect(err).To(MatchError(actionerror.NotLoggedInError{
				BinaryName: binaryName,
			}))
		})
	})

	Context("when the user is logged in", func() {
		BeforeEach(func() {
			fakeConfig.AccessTokenReturns("some-access-token")
			fakeConfig.RefreshTokenReturns("some-refresh-token")
		})

		DescribeTable("targeting org check",
			func(isOrgTargeted bool, checkForTargeted bool, expectedError error) {
				fakeConfig.HasTargetedOrganizationReturns(isOrgTargeted)

				err := actor.CheckTarget(checkForTargeted, false)

				if expectedError != nil {
					Expect(err).To(MatchError(expectedError))
				} else {
					Expect(err).ToNot(HaveOccurred())
				}
			},

			Entry("it returns an error", false, true, actionerror.NoOrganizationTargetedError{BinaryName: "faceman"}),
			Entry("it does not return an error", false, false, nil),
			Entry("it does not return an error", true, false, nil),
			Entry("it does not return an error", true, true, nil),
		)

		Context("when the organization is targeted", func() {
			BeforeEach(func() {
				fakeConfig.HasTargetedOrganizationReturns(true)
			})

			DescribeTable("targeting space check",
				func(isSpaceTargeted bool, checkForTargeted bool, expectedError error) {
					fakeConfig.HasTargetedSpaceReturns(isSpaceTargeted)

					err := actor.CheckTarget(true, checkForTargeted)

					if expectedError != nil {
						Expect(err).To(MatchError(expectedError))
					} else {
						Expect(err).ToNot(HaveOccurred())
					}
				},

				Entry("it returns an error", false, true, actionerror.NoSpaceTargetedError{BinaryName: "faceman"}),
				Entry("it does not return an error", false, false, nil),
				Entry("it does not return an error", true, false, nil),
				Entry("it does not return an error", true, true, nil),
			)
		})
	})
})
