package v3_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v3"
	"code.cloudfoundry.org/cli/command/v3/v3fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("share-service Command", func() {
	var (
		cmd             v3.V3ShareServiceCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v3fakes.FakeShareServiceActor
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v3fakes.FakeShareServiceActor)

		cmd = v3.V3ShareServiceCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

		cmd.RequiredArgs.ServiceInstance = "some-service-instance"

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		fakeActor.CloudControllerAPIVersionReturns(ccversion.MinVersionShareServiceV3)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	Context("when the API version is below the minimum", func() {
		BeforeEach(func() {
			fakeActor.CloudControllerAPIVersionReturns("0.0.0")
		})

		It("returns a MinimumAPIVersionNotMetError", func() {
			Expect(executeErr).To(MatchError(translatableerror.MinimumAPIVersionNotMetError{
				CurrentVersion: "0.0.0",
				MinimumVersion: ccversion.MinVersionShareServiceV3,
			}))
		})
	})

	Context("when checking target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: binaryName}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeTrue())
			Expect(checkTargetedSpace).To(BeTrue())
		})
	})

	Context("when the user is logged in, and a space and org are targeted", func() {
		BeforeEach(func() {
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				GUID: "some-org-guid",
				Name: "some-org",
			})
			fakeConfig.TargetedSpaceReturns(configv3.Space{
				GUID: "some-space-guid",
				Name: "some-space",
			})
		})

		Context("when getting the current user returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("get current user error")
				fakeConfig.CurrentUserReturns(
					configv3.User{},
					expectedErr)
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError(expectedErr))
			})
		})

		Context("when the current user is set correctly", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserReturns(
					configv3.User{Name: "some-user"},
					nil)
			})

			Context("when using the currently targeted org", func() {
				Context("when the sharing is successful", func() {
					BeforeEach(func() {
						cmd.SpaceName = "some-space"
						fakeActor.ShareServiceInstanceInSpaceByOrganizationAndSpaceNameReturns(
							v3action.Warnings{"share-service-warning"},
							nil)
					})

					It("shares the service instance with the provided space and displays all warnings", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say("Sharing service instance some-service-instance into org some-org / space some-space as some-user\\.\\.\\."))
						Expect(testUI.Out).To(Say("OK"))
						Expect(testUI.Err).To(Say("share-service-warning"))

						Expect(fakeActor.ShareServiceInstanceInSpaceByOrganizationAndSpaceNameCallCount()).To(Equal(1))
						serviceInstanceNameArg, sourceSpaceGUIDArg, orgGUIDArg, spaceNameArg := fakeActor.ShareServiceInstanceInSpaceByOrganizationAndSpaceNameArgsForCall(0)
						Expect(serviceInstanceNameArg).To(Equal("some-service-instance"))
						Expect(sourceSpaceGUIDArg).To(Equal("some-space-guid"))
						Expect(orgGUIDArg).To(Equal("some-org-guid"))
						Expect(spaceNameArg).To(Equal("some-space"))
					})
				})

				Context("when the sharing is unsuccessful", func() {
					BeforeEach(func() {
						cmd.SpaceName = "some-space"
						fakeActor.ShareServiceInstanceInSpaceByOrganizationAndSpaceNameReturns(
							v3action.Warnings{"share-service-warning"},
							errors.New("sharing failed"))
					})

					It("does not share the service instance with the provided space and displays all warnings", func() {
						Expect(executeErr).To(MatchError("sharing failed"))

						Expect(testUI.Out).ToNot(Say("OK"))
						Expect(testUI.Err).To(Say("share-service-warning"))
					})
				})
			})

			Context("when using a specified org", func() {
				Context("when the sharing is successful", func() {
					BeforeEach(func() {
						cmd.SpaceName = "some-space"
						cmd.OrgName = "some-other-org"
						fakeActor.ShareServiceInstanceInSpaceByOrganizationNameAndSpaceNameReturns(
							v3action.Warnings{"share-service-warning"},
							nil)
					})

					It("shares the service instance with the provided space and org and displays all warnings", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say("Sharing service instance some-service-instance into org some-other-org / space some-space as some-user\\.\\.\\."))
						Expect(testUI.Out).To(Say("OK"))
						Expect(testUI.Err).To(Say("share-service-warning"))

						Expect(fakeActor.ShareServiceInstanceInSpaceByOrganizationNameAndSpaceNameCallCount()).To(Equal(1))
						serviceInstanceNameArg, sourceSpaceGUID, orgName, spaceNameArg := fakeActor.ShareServiceInstanceInSpaceByOrganizationNameAndSpaceNameArgsForCall(0)
						Expect(serviceInstanceNameArg).To(Equal("some-service-instance"))
						Expect(sourceSpaceGUID).To(Equal("some-space-guid"))
						Expect(orgName).To(Equal("some-other-org"))
						Expect(spaceNameArg).To(Equal("some-space"))
					})
				})

				Context("when the sharing is unsuccessful", func() {
					BeforeEach(func() {
						cmd.SpaceName = "some-space"
						cmd.OrgName = "some-other-org"
						fakeActor.ShareServiceInstanceInSpaceByOrganizationNameAndSpaceNameReturns(
							v3action.Warnings{"share-service-warning"},
							errors.New("sharing failed"))
					})

					It("does not share the service instance with the provided space and displays all warnings", func() {
						Expect(executeErr).To(MatchError("sharing failed"))

						Expect(testUI.Out).ToNot(Say("OK"))
						Expect(testUI.Err).To(Say("share-service-warning"))
					})
				})
			})
		})
	})
})
