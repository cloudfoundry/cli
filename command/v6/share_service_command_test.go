package v6_test

import (
	"errors"

	"code.cloudfoundry.org/cli/command/translatableerror"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v2v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v6"
	"code.cloudfoundry.org/cli/command/v6/v6fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("share-service Command", func() {
	var (
		cmd             ShareServiceCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v6fakes.FakeShareServiceActor
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v6fakes.FakeShareServiceActor)

		cmd = ShareServiceCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

		cmd.RequiredArgs.ServiceInstance = "some-service-instance"
		cmd.SpaceName = "some-space"

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		fakeActor.CloudControllerV3APIVersionReturns(ccversion.MinSupportedV3ClientVersion)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("the API version is below the minimum", func() {
		currentVersion := "3.0.1"
		BeforeEach(func() {
			fakeActor.CloudControllerV3APIVersionReturns(currentVersion)
		})

		It("returns a MinimumAPIVersionNotMetError", func() {
			Expect(executeErr).To(MatchError(translatableerror.MinimumCFAPIVersionNotMetError{
				CurrentVersion: currentVersion,
				MinimumVersion: ccversion.MinSupportedV3ClientVersion,
			}))
		})
	})

	When("the environment is not correctly setup", func() {
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

	When("the user is logged in, and a space and org are targeted", func() {
		BeforeEach(func() {
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				GUID: "some-org-guid",
				Name: "some-org",
			})
			fakeConfig.TargetedSpaceReturns(configv3.Space{
				GUID: "some-space-guid",
			})
		})

		When("an error occurs getting the current user", func() {
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

		When("no errors occur getting the current user", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserReturns(
					configv3.User{Name: "some-user"},
					nil)
			})

			When("'-o' (org name) is not provided", func() {
				When("no errors occur sharing the service instance", func() {
					BeforeEach(func() {
						fakeActor.ShareServiceInstanceToSpaceNameByNameAndSpaceAndOrganizationReturns(
							v2v3action.Warnings{"share-service-warning"},
							nil)
					})

					It("shares the service instance with the provided space and displays all warnings", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say(`Sharing service instance some-service-instance into org some-org / space some-space as some-user\.\.\.`))
						Expect(testUI.Out).To(Say("OK"))

						Expect(testUI.Err).To(Say("share-service-warning"))

						Expect(fakeActor.ShareServiceInstanceToSpaceNameByNameAndSpaceAndOrganizationCallCount()).To(Equal(1))
						spaceNameArg, serviceInstanceNameArg, sourceSpaceGUIDArg, orgGUIDArg := fakeActor.ShareServiceInstanceToSpaceNameByNameAndSpaceAndOrganizationArgsForCall(0)
						Expect(spaceNameArg).To(Equal("some-space"))
						Expect(serviceInstanceNameArg).To(Equal("some-service-instance"))
						Expect(sourceSpaceGUIDArg).To(Equal("some-space-guid"))
						Expect(orgGUIDArg).To(Equal("some-org-guid"))
					})
				})

				When("an error occurs sharing the service instance", func() {
					var expectedErr error

					BeforeEach(func() {
						expectedErr = errors.New("sharing failed")
						fakeActor.ShareServiceInstanceToSpaceNameByNameAndSpaceAndOrganizationReturns(
							v2v3action.Warnings{"share-service-warning"},
							expectedErr)
					})

					It("returns the error, and displays all warnings", func() {
						Expect(executeErr).To(MatchError(expectedErr))

						Expect(testUI.Out).ToNot(Say("OK"))
						Expect(testUI.Err).To(Say("share-service-warning"))
					})
				})

				When("the service instance is not shareable", func() {
					var expectedErr error

					BeforeEach(func() {
						expectedErr = actionerror.ServiceInstanceNotShareableError{
							FeatureFlagEnabled:          true,
							ServiceBrokerSharingEnabled: false}
						fakeActor.ShareServiceInstanceToSpaceNameByNameAndSpaceAndOrganizationReturns(
							v2v3action.Warnings{"share-service-instance-warning"},
							expectedErr)
					})

					It("returns ServiceInstanceNotShareableError and displays all warnings", func() {
						Expect(executeErr).To(MatchError(expectedErr))

						Expect(testUI.Out).ToNot(Say("OK"))
						Expect(testUI.Err).To(Say("share-service-instance-warning"))
					})
				})

				When("the service instance is already shared to the space", func() {
					BeforeEach(func() {
						fakeActor.ShareServiceInstanceToSpaceNameByNameAndSpaceAndOrganizationReturns(
							v2v3action.Warnings{"share-service-warning"},
							actionerror.ServiceInstanceAlreadySharedError{})
					})

					It("does not return an error and displays all warnings", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say("Service instance some-service-instance is already shared with that space."))
						Expect(testUI.Out).To(Say("OK"))

						Expect(testUI.Err).To(Say("share-service-warning"))
					})
				})
			})

			When("-o (org name) is provided", func() {
				BeforeEach(func() {
					cmd.OrgName = "some-other-org"
				})

				When("no errors occur sharing the service instance", func() {
					BeforeEach(func() {
						fakeActor.ShareServiceInstanceToSpaceNameByNameAndSpaceAndOrganizationNameReturns(
							v2v3action.Warnings{"share-service-warning"},
							nil)
					})

					It("shares the service instance with the provided space and org and displays all warnings", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say(`Sharing service instance some-service-instance into org some-other-org / space some-space as some-user\.\.\.`))
						Expect(testUI.Out).To(Say("OK"))

						Expect(testUI.Err).To(Say("share-service-warning"))

						Expect(fakeActor.ShareServiceInstanceToSpaceNameByNameAndSpaceAndOrganizationNameCallCount()).To(Equal(1))
						spaceNameArg, serviceInstanceNameArg, sourceSpaceGUID, orgName := fakeActor.ShareServiceInstanceToSpaceNameByNameAndSpaceAndOrganizationNameArgsForCall(0)
						Expect(spaceNameArg).To(Equal("some-space"))
						Expect(serviceInstanceNameArg).To(Equal("some-service-instance"))
						Expect(sourceSpaceGUID).To(Equal("some-space-guid"))
						Expect(orgName).To(Equal("some-other-org"))
					})
				})

				When("an error occurs sharing the service instance", func() {
					var expectedErr error

					BeforeEach(func() {
						expectedErr = errors.New("sharing failed")
						fakeActor.ShareServiceInstanceToSpaceNameByNameAndSpaceAndOrganizationNameReturns(
							v2v3action.Warnings{"share-service-warning"},
							expectedErr)
					})

					It("returns the error and displays all warnings", func() {
						Expect(executeErr).To(MatchError(expectedErr))

						Expect(testUI.Out).ToNot(Say("OK"))
						Expect(testUI.Err).To(Say("share-service-warning"))
					})
				})

				When("the service instance is not shareable", func() {
					var expectedErr error

					BeforeEach(func() {
						expectedErr = actionerror.ServiceInstanceNotShareableError{
							FeatureFlagEnabled:          false,
							ServiceBrokerSharingEnabled: true}
						fakeActor.ShareServiceInstanceToSpaceNameByNameAndSpaceAndOrganizationNameReturns(
							v2v3action.Warnings{"share-service-instance-warning"},
							expectedErr)
					})

					It("returns ServiceInstanceNotShareableError and displays all warnings", func() {
						Expect(executeErr).To(MatchError(expectedErr))

						Expect(testUI.Out).ToNot(Say("OK"))
						Expect(testUI.Err).To(Say("share-service-instance-warning"))
					})
				})

				When("the service instance is already shared to the space", func() {
					BeforeEach(func() {
						fakeActor.ShareServiceInstanceToSpaceNameByNameAndSpaceAndOrganizationNameReturns(
							v2v3action.Warnings{"share-service-warning"},
							actionerror.ServiceInstanceAlreadySharedError{})
					})

					It("does not return an error and displays all warnings", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say("Service instance some-service-instance is already shared with that space."))
						Expect(testUI.Out).To(Say("OK"))

						Expect(testUI.Err).To(Say("share-service-warning"))
					})
				})
			})
		})
	})
})
