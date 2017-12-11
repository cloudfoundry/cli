package v3_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v2action"
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

var _ = Describe("unshare-service Command", func() {
	var (
		cmd             v3.V3UnshareServiceCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v3fakes.FakeUnshareServiceActor
		fakeActorV2     *v3fakes.FakeServiceInstanceSharedToActorV2
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v3fakes.FakeUnshareServiceActor)
		fakeActorV2 = new(v3fakes.FakeServiceInstanceSharedToActorV2)

		cmd = v3.V3UnshareServiceCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
			ActorV2:     fakeActorV2,
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

			Context("when looking up the shared-to space guid returns an error", func() {
				BeforeEach(func() {
					fakeActorV2.GetSharedToSpaceGUIDReturns(
						"",
						v2action.Warnings{"get-shared-to-space-guid-warning"},
						errors.New("an error"))
				})

				It("returns the error and displays all warnings", func() {
					Expect(executeErr).To(MatchError("an error"))
					Expect(testUI.Err).To(Say("get-shared-to-space-guid-warning"))
				})
			})

			Context("when using the currently targeted org", func() {
				BeforeEach(func() {
					cmd.SpaceName = "some-space"
					fakeActorV2.GetSharedToSpaceGUIDReturns(
						"shared-to-space-guid",
						v2action.Warnings{"get-shared-to-space-guid-warning"},
						nil)
				})

				It("calls GetSharedToSpaceGUID with the correct parameters", func() {
					Expect(fakeActorV2.GetSharedToSpaceGUIDCallCount()).To(Equal(1))
					serviceInstanceNameArg, sourceSpaceGUIDArg, sharedToOrgNameArg, sharedToSpaceNameArg := fakeActorV2.GetSharedToSpaceGUIDArgsForCall(0)
					Expect(serviceInstanceNameArg).To(Equal("some-service-instance"))
					Expect(sourceSpaceGUIDArg).To(Equal("some-space-guid"))
					Expect(sharedToOrgNameArg).To(Equal("some-org"))
					Expect(sharedToSpaceNameArg).To(Equal("some-space"))
				})

				Context("when the unsharing is successful", func() {
					BeforeEach(func() {
						fakeActor.UnshareServiceInstanceFromSpaceReturns(
							v3action.Warnings{"unshare-service-warning"},
							nil)
					})

					It("unshares the service instance with the provided space and displays all warnings", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say("Unsharing service instance some-service-instance from org some-org / space some-space as some-user\\.\\.\\."))
						Expect(testUI.Out).To(Say("OK"))
						Expect(testUI.Err).To(Say("get-shared-to-space-guid-warning"))
						Expect(testUI.Err).To(Say("unshare-service-warning"))

						Expect(fakeActor.UnshareServiceInstanceFromSpaceCallCount()).To(Equal(1))
						serviceInstanceNameArg, sourceSpaceGUIDArg, sharedToSpaceGUIDArg := fakeActor.UnshareServiceInstanceFromSpaceArgsForCall(0)
						Expect(serviceInstanceNameArg).To(Equal("some-service-instance"))
						Expect(sourceSpaceGUIDArg).To(Equal("some-space-guid"))
						Expect(sharedToSpaceGUIDArg).To(Equal("shared-to-space-guid"))
					})
				})

				Context("when the unsharing is unsuccessful", func() {
					BeforeEach(func() {
						fakeActor.UnshareServiceInstanceFromSpaceReturns(
							v3action.Warnings{"unshare-service-warning"},
							errors.New("unsharing failed"))
					})

					It("does not unshare the service instance with the provided space and displays all warnings", func() {
						Expect(executeErr).To(MatchError("unsharing failed"))

						Expect(testUI.Err).To(Say("get-shared-to-space-guid-warning"))
						Expect(testUI.Err).To(Say("unshare-service-warning"))
					})
				})
			})

			Context("when using a specified org", func() {
				BeforeEach(func() {
					cmd.SpaceName = "some-space"
					cmd.OrgName = "some-other-org"
					fakeActorV2.GetSharedToSpaceGUIDReturns(
						"shared-to-space-guid",
						v2action.Warnings{"get-shared-to-space-guid-warning"},
						nil)
				})

				It("calls GetSharedToSpaceGUID with the correct parameters", func() {
					Expect(fakeActorV2.GetSharedToSpaceGUIDCallCount()).To(Equal(1))
					serviceInstanceNameArg, sourceSpaceGUIDArg, sharedToOrgNameArg, sharedToSpaceNameArg := fakeActorV2.GetSharedToSpaceGUIDArgsForCall(0)
					Expect(serviceInstanceNameArg).To(Equal("some-service-instance"))
					Expect(sourceSpaceGUIDArg).To(Equal("some-space-guid"))
					Expect(sharedToOrgNameArg).To(Equal("some-other-org"))
					Expect(sharedToSpaceNameArg).To(Equal("some-space"))
				})

				Context("when the unsharing is successful", func() {
					BeforeEach(func() {
						fakeActor.UnshareServiceInstanceFromSpaceReturns(
							v3action.Warnings{"unshare-service-warning"},
							nil)
					})

					It("unshares the service instance with the provided space and org and displays all warnings", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say("Unsharing service instance some-service-instance from org some-other-org / space some-space as some-user\\.\\.\\."))
						Expect(testUI.Out).To(Say("OK"))
						Expect(testUI.Err).To(Say("get-shared-to-space-guid-warning"))
						Expect(testUI.Err).To(Say("unshare-service-warning"))

						Expect(fakeActor.UnshareServiceInstanceFromSpaceCallCount()).To(Equal(1))
						serviceInstanceNameArg, sourceSpaceGUID, sharedToSpaceGUID := fakeActor.UnshareServiceInstanceFromSpaceArgsForCall(0)
						Expect(serviceInstanceNameArg).To(Equal("some-service-instance"))
						Expect(sourceSpaceGUID).To(Equal("some-space-guid"))
						Expect(sharedToSpaceGUID).To(Equal("shared-to-space-guid"))
					})
				})

				Context("when the unsharing is unsuccessful", func() {
					BeforeEach(func() {
						fakeActor.UnshareServiceInstanceFromSpaceReturns(
							v3action.Warnings{"unshare-service-warning"},
							errors.New("unsharing failed"))
					})

					It("does not unshare the service instance from the provided space and displays all warnings", func() {
						Expect(executeErr).To(MatchError("unsharing failed"))

						Expect(testUI.Err).To(Say("get-shared-to-space-guid-warning"))
						Expect(testUI.Err).To(Say("unshare-service-warning"))
					})
				})
			})
		})
	})
})
