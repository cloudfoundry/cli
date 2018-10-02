package sharedaction_test

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/sharedaction/sharedactionfakes"
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Checking target", func() {
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

	Describe("CheckTarget", func() {
		When("the user is not logged in", func() {
			It("returns an error", func() {
				err := actor.CheckTarget(false, false)
				Expect(err).To(MatchError(actionerror.NotLoggedInError{
					BinaryName: binaryName,
				}))
			})
		})

		When("the user is logged in", func() {
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

			When("the organization is targeted", func() {
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

	Describe("RequireCurrentUser", func() {
		var (
			user string
			err  error
		)

		JustBeforeEach(func() {
			user, err = actor.RequireCurrentUser()
		})

		When("access token and refresh token are empty", func() {
			BeforeEach(func() {
				fakeConfig.AccessTokenReturns("")
				fakeConfig.RefreshTokenReturns("")
			})

			It("returns a NotLoggedInError with binary name", func() {
				Expect(err).To(MatchError(actionerror.NotLoggedInError{
					BinaryName: binaryName,
				}))
			})
		})

		When("access token is available", func() {
			BeforeEach(func() {
				fakeConfig.AccessTokenReturns("some-access-token")
			})

			When("getting current user returns an error", func() {
				BeforeEach(func() {
					fakeConfig.CurrentUserNameReturns("", errors.New("get-user-error"))
				})

				It("returns the error", func() {
					Expect(err).To(MatchError("get-user-error"))
				})
			})

			When("getting current user succeeds", func() {
				BeforeEach(func() {
					fakeConfig.CurrentUserNameReturns("some-user", nil)
				})

				It("returns the user", func() {
					Expect(err).To(BeNil())
					Expect(user).To(Equal("some-user"))
				})
			})
		})

		When("refresh token is available", func() {
			BeforeEach(func() {
				fakeConfig.RefreshTokenReturns("some-refresh-token")
			})

			When("getting current user returns an error", func() {
				BeforeEach(func() {
					fakeConfig.CurrentUserNameReturns("", errors.New("get-user-error"))
				})

				It("returns the error", func() {
					Expect(err).To(MatchError("get-user-error"))
				})
			})

			When("getting current user succeeds", func() {
				BeforeEach(func() {
					fakeConfig.CurrentUserNameReturns("some-user", nil)
				})

				It("returns the user object", func() {
					Expect(err).To(BeNil())
					Expect(user).To(Equal("some-user"))
				})
			})
		})
	})

	Describe("RequireTargetedOrg", func() {
		var (
			org string
			err error
		)

		JustBeforeEach(func() {
			org, err = actor.RequireTargetedOrg()
		})

		When("no org is targeted", func() {
			BeforeEach(func() {
				fakeConfig.HasTargetedOrganizationReturns(false)
			})

			It("returns a NoOrganizationTargetedError with the binary name", func() {
				Expect(err).To(MatchError(actionerror.NoOrganizationTargetedError{
					BinaryName: binaryName,
				}))
				Expect(org).To(BeEmpty())
			})
		})

		When("an org is targeted", func() {
			BeforeEach(func() {
				fakeConfig.HasTargetedOrganizationReturns(true)
				fakeConfig.TargetedOrganizationNameReturns("some-name")
			})

			It("returns the org name", func() {
				Expect(err).To(BeNil())
				Expect(org).To(Equal("some-name"))
			})
		})
	})
})
