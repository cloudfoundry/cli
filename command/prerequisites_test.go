package command_test

import (
	. "code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/commandfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("CheckTarget", func() {
	var (
		binaryName string
		fakeConfig *commandfakes.FakeConfig
	)

	BeforeEach(func() {
		binaryName = "faceman"
		fakeConfig = new(commandfakes.FakeConfig)
		fakeConfig.BinaryNameReturns(binaryName)
	})

	Context("when the api endpoint is set", func() {
		BeforeEach(func() {
			fakeConfig.TargetReturns("some-url")
		})

		Context("when the user is not logged in", func() {
			It("returns an error", func() {
				err := CheckTarget(fakeConfig, false, false)
				Expect(err).To(MatchError(NotLoggedInError{
					BinaryName: binaryName,
				}))
			})
		})

		Context("when the user is logged in", func() {
			BeforeEach(func() {
				fakeConfig.AccessTokenReturns("some-access-token")
				fakeConfig.RefreshTokenReturns("some-refresh-token")
			})

			DescribeTable("targeting organization check",
				func(isOrgTargeted bool, checkForTargeted bool, expectedError error) {
					fakeConfig.HasTargetedOrganizationReturns(isOrgTargeted)

					err := CheckTarget(fakeConfig, checkForTargeted, false)

					if expectedError != nil {
						Expect(err).To(MatchError(expectedError))
					} else {
						Expect(err).ToNot(HaveOccurred())
					}
				},

				Entry("it returns an error", false, true, NoTargetedOrgError{BinaryName: "faceman"}),
				Entry("it does not return an error", false, false, nil),
				Entry("it does not return an error", true, false, nil),
				Entry("it does not return an error", true, true, nil),
			)

			Context("when the organization is targeted", func() {
				DescribeTable("targeting space check",
					func(isSpaceTargeted bool, checkForTargeted bool, expectedError error) {
						fakeConfig.HasTargetedOrganizationReturns(true)
						fakeConfig.HasTargetedSpaceReturns(isSpaceTargeted)

						err := CheckTarget(fakeConfig, true, checkForTargeted)

						if expectedError != nil {
							Expect(err).To(MatchError(expectedError))
						} else {
							Expect(err).ToNot(HaveOccurred())
						}
					},

					Entry("it returns an error", false, true, NoTargetedSpaceError{BinaryName: "faceman"}),
					Entry("it does not return an error", false, false, nil),
					Entry("it does not return an error", true, false, nil),
					Entry("it does not return an error", true, true, nil),
				)
			})
		})
	})
})
