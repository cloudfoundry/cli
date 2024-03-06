package v6_test

import (
	"errors"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/command/translatableerror"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v2v3action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v6"
	"code.cloudfoundry.org/cli/command/v6/v6fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("unshare-service Command", func() {
	var (
		cmd             UnshareServiceCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v6fakes.FakeUnshareServiceActor
		input           *Buffer
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v6fakes.FakeUnshareServiceActor)

		cmd = UnshareServiceCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

		cmd.RequiredArgs.ServiceInstance = "some-service-instance"

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		fakeActor.CloudControllerV3APIVersionReturns(ccversion.MinSupportedV3ClientVersion)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("the API version is below the minimum", func() {
		olderCurrentVersion := "3.0.1"
		BeforeEach(func() {
			fakeActor.CloudControllerV3APIVersionReturns(olderCurrentVersion)
		})

		It("returns a MinimumAPIVersionNotMetError", func() {
			Expect(executeErr).To(MatchError(translatableerror.MinimumCFAPIVersionNotMetError{
				CurrentVersion: olderCurrentVersion,
				MinimumVersion: ccversion.MinSupportedV3ClientVersion,
			}))
		})
	})

	When("checking target fails", func() {
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
		const (
			sharedToSpaceName = "some-shared-to-space"
			expectedWarning   = `WARNING: Unsharing this service instance will remove any existing bindings originating from the service instance in the space \"` + sharedToSpaceName + `\". This could cause apps to stop working.`
		)

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
				Expect(fakeConfig.CurrentUserCallCount()).To(Equal(1))
			})
		})

		When("no errors occur getting the current user", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserReturns(
					configv3.User{Name: "some-user"},
					nil)
			})

			When("the '-f' flag is provided (to force non-prompting)", func() {
				BeforeEach(func() {
					cmd.Force = true
				})

				When("the '-o' flag is NOT provided (when we want to unshare a space in the currently targeted org)", func() {
					BeforeEach(func() {
						cmd.SharedToSpaceName = "some-shared-to-space"
					})

					When("no errors occur unsharing the service instance", func() {
						BeforeEach(func() {
							fakeActor.UnshareServiceInstanceFromOrganizationNameAndSpaceNameByNameAndSpaceReturns(
								v2v3action.Warnings{"unshare-service-warning"},
								nil)
						})

						It("unshares the service instance from the currently targetd org and provided space name, and displays all warnings", func() {
							Expect(executeErr).ToNot(HaveOccurred())

							Expect(testUI.Out).To(Say(`Unsharing service instance some-service-instance from org some-org / space some-shared-to-space as some-user\.\.\.`))
							Expect(testUI.Out).To(Say("OK"))

							Expect(testUI.Err).ToNot(Say("WARNING"))
							Expect(testUI.Err).To(Say("unshare-service-warning"))

							Expect(fakeActor.UnshareServiceInstanceFromOrganizationNameAndSpaceNameByNameAndSpaceCallCount()).To(Equal(1))
							sharedToOrgNameArg, sharedToSpaceNameArg, serviceInstanceNameArg, currentlyTargetedSpaceGUIDArg := fakeActor.UnshareServiceInstanceFromOrganizationNameAndSpaceNameByNameAndSpaceArgsForCall(0)
							Expect(sharedToOrgNameArg).To(Equal("some-org"))
							Expect(sharedToSpaceNameArg).To(Equal("some-shared-to-space"))
							Expect(serviceInstanceNameArg).To(Equal("some-service-instance"))
							Expect(currentlyTargetedSpaceGUIDArg).To(Equal("some-space-guid"))
						})
					})

					When("an error occurs unsharing the service instance", func() {
						var expectedErr error

						BeforeEach(func() {
							expectedErr = errors.New("unshare error")
							fakeActor.UnshareServiceInstanceFromOrganizationNameAndSpaceNameByNameAndSpaceReturns(
								v2v3action.Warnings{"unshare-service-warning"},
								expectedErr)
						})

						It("returns the error and displays all warnings", func() {
							Expect(executeErr).To(MatchError(expectedErr))
							Expect(testUI.Err).To(Say("unshare-service-warning"))
						})
					})

					When("the service instance is not shared to the space we want to unshare from", func() {
						BeforeEach(func() {
							fakeActor.UnshareServiceInstanceFromOrganizationNameAndSpaceNameByNameAndSpaceReturns(
								v2v3action.Warnings{"unshare-service-warning"},
								actionerror.ServiceInstanceNotSharedToSpaceError{ServiceInstanceName: "some-service-instance"})
						})

						It("does not return an error, displays that the service instance is not shared and displays all warnings", func() {
							Expect(executeErr).ToNot(HaveOccurred())
							Expect(testUI.Out).To(Say(`Service instance some-service-instance is not shared with space some-shared-to-space in organization some-org\.`))
							Expect(testUI.Out).To(Say("OK"))
							Expect(testUI.Err).To(Say("unshare-service-warning"))
						})
					})
				})

				When("the '-o' flag is provided (when the space we want to unshare does not belong to the currently targeted org)", func() {
					BeforeEach(func() {
						cmd.SharedToSpaceName = "some-other-space"
						cmd.SharedToOrgName = "some-other-org"

						fakeActor.UnshareServiceInstanceFromOrganizationNameAndSpaceNameByNameAndSpaceReturns(
							v2v3action.Warnings{"unshare-service-warning"},
							nil)
					})

					It("performs the unshare with the provided org name", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say(`Unsharing service instance some-service-instance from org some-other-org / space some-other-space as some-user\.\.\.`))
						Expect(testUI.Out).To(Say("OK"))

						Expect(testUI.Err).ToNot(Say("WARNING"))
						Expect(testUI.Err).To(Say("unshare-service-warning"))

						Expect(fakeActor.UnshareServiceInstanceFromOrganizationNameAndSpaceNameByNameAndSpaceCallCount()).To(Equal(1))
						sharedToOrgNameArg, sharedToSpaceNameArg, serviceInstanceNameArg, currentlyTargetedSpaceGUIDArg := fakeActor.UnshareServiceInstanceFromOrganizationNameAndSpaceNameByNameAndSpaceArgsForCall(0)
						Expect(sharedToOrgNameArg).To(Equal("some-other-org"))
						Expect(sharedToSpaceNameArg).To(Equal("some-other-space"))
						Expect(serviceInstanceNameArg).To(Equal("some-service-instance"))
						Expect(currentlyTargetedSpaceGUIDArg).To(Equal("some-space-guid"))
					})
				})
			})

			When("the -f flag is NOT provided", func() {
				BeforeEach(func() {
					cmd.Force = false
					cmd.SharedToSpaceName = sharedToSpaceName
				})

				When("the user inputs yes", func() {
					BeforeEach(func() {
						_, err := input.Write([]byte("y\n"))
						Expect(err).ToNot(HaveOccurred())

						fakeActor.UnshareServiceInstanceFromOrganizationNameAndSpaceNameByNameAndSpaceReturns(
							v2v3action.Warnings{"unshare-service-warning"},
							nil)
					})

					It("unshares the service instance and displays all warnings", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say(`Really unshare the service instance\? \[yN\]`))
						Expect(testUI.Out).To(Say(`Unsharing service instance some-service-instance from org some-org / space some-shared-to-space as some-user\.\.\.`))
						Expect(testUI.Out).To(Say("OK"))

						Expect(testUI.Err).To(Say(expectedWarning))
						Expect(testUI.Err).To(Say("unshare-service-warning"))

						Expect(fakeActor.UnshareServiceInstanceFromOrganizationNameAndSpaceNameByNameAndSpaceCallCount()).To(Equal(1))
						sharedToOrgNameArg, sharedToSpaceNameArg, serviceInstanceNameArg, currentlyTargetedSpaceGUIDArg := fakeActor.UnshareServiceInstanceFromOrganizationNameAndSpaceNameByNameAndSpaceArgsForCall(0)
						Expect(sharedToOrgNameArg).To(Equal("some-org"))
						Expect(sharedToSpaceNameArg).To(Equal("some-shared-to-space"))
						Expect(serviceInstanceNameArg).To(Equal("some-service-instance"))
						Expect(currentlyTargetedSpaceGUIDArg).To(Equal("some-space-guid"))
					})
				})

				When("the user inputs no", func() {
					BeforeEach(func() {
						_, err := input.Write([]byte("n\n"))
						Expect(err).ToNot(HaveOccurred())
					})

					It("cancels the unshared", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say(`Really unshare the service instance\? \[yN\]`))
						Expect(testUI.Out).To(Say("Unshare cancelled"))

						Expect(testUI.Err).To(Say(expectedWarning))

						Expect(fakeActor.UnshareServiceInstanceFromOrganizationNameAndSpaceNameByNameAndSpaceCallCount()).To(Equal(0))
					})
				})

				When("the user chooses the default", func() {
					BeforeEach(func() {
						_, err := input.Write([]byte("\n"))
						Expect(err).ToNot(HaveOccurred())
					})

					It("cancels the unshare", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say(`Really unshare the service instance\? \[yN\]`))
						Expect(testUI.Out).To(Say("Unshare cancelled"))

						Expect(testUI.Err).To(Say(expectedWarning))

						Expect(fakeActor.UnshareServiceInstanceFromOrganizationNameAndSpaceNameByNameAndSpaceCallCount()).To(Equal(0))
					})
				})

				When("the user input is invalid", func() {
					BeforeEach(func() {
						_, err := input.Write([]byte("e\n\n"))
						Expect(err).ToNot(HaveOccurred())
					})

					It("prompts for unshare confirmation again", func() {
						Expect(executeErr).NotTo(HaveOccurred())

						Expect(testUI.Out).To(Say(`Really unshare the service instance\? \[yN\]`))
						Expect(testUI.Out).To(Say(`invalid input \(not y, n, yes, or no\)`))
						Expect(testUI.Out).To(Say(`Really unshare the service instance\? \[yN\]`))

						Expect(testUI.Err).To(Say(expectedWarning))

						Expect(fakeActor.UnshareServiceInstanceFromOrganizationNameAndSpaceNameByNameAndSpaceCallCount()).To(Equal(0))
					})
				})
			})
		})
	})
})
