package v2_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	. "code.cloudfoundry.org/cli/command/v2"
	"code.cloudfoundry.org/cli/command/v2/v2fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("unbind-security-group Command", func() {
	var (
		cmd             UnbindSecurityGroupCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v2fakes.FakeUnbindSecurityGroupActor
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v2fakes.FakeUnbindSecurityGroupActor)

		cmd = UnbindSecurityGroupCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		fakeConfig.ExperimentalReturns(true)

		fakeConfig.CurrentUserReturns(
			configv3.User{Name: "some-user"},
			nil)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	Context("when getting the current user fails", func() {
		var expectedErr error

		BeforeEach(func() {
			expectedErr = errors.New("getting user failed")
			fakeConfig.CurrentUserReturns(configv3.User{}, expectedErr)
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(expectedErr))
		})
	})

	Context("when checking target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NoOrganizationTargetedError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NoOrganizationTargetedError{BinaryName: "faceman"}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeTrue())
			Expect(checkTargetedSpace).To(BeTrue())
		})
	})

	Context("when lifecycle is 'some-lifecycle'", func() {
		// By this point in execution, Goflags will have filtered any invalid
		// lifecycle phase.  We use 'some-lifecycle' to test that the command
		// merely passes the value presented by Goflags.
		BeforeEach(func() {
			cmd.Lifecycle = flag.SecurityGroupLifecycle("some-lifecycle")
		})

		Context("when only the security group is provided", func() {
			BeforeEach(func() {
				cmd.RequiredArgs.SecurityGroupName = "some-security-group"
			})

			Context("when org and space are targeted", func() {
				BeforeEach(func() {
					fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "some-org"})
					fakeConfig.TargetedSpaceReturns(configv3.Space{Name: "some-space", GUID: "some-space-guid"})
					fakeActor.UnbindSecurityGroupByNameAndSpaceReturns(
						v2action.Warnings{"unbind warning"},
						nil)
				})

				It("unbinds the security group from the targeted space", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say("Unbinding security group %s from org %s / space %s as %s\\.\\.\\.", "some-security-group", "some-org", "some-space", "some-user"))
					Expect(testUI.Out).To(Say("OK\n\n"))
					Expect(testUI.Out).To(Say("TIP: Changes require an app restart \\(for running\\) or restage \\(for staging\\) to apply to existing applications\\."))
					Expect(testUI.Err).To(Say("unbind warning"))

					Expect(fakeConfig.TargetedOrganizationCallCount()).To(Equal(1))
					Expect(fakeConfig.TargetedSpaceCallCount()).To(Equal(1))
					Expect(fakeActor.UnbindSecurityGroupByNameAndSpaceCallCount()).To(Equal(1))
					securityGroupName, spaceGUID, lifecycle := fakeActor.UnbindSecurityGroupByNameAndSpaceArgsForCall(0)
					Expect(securityGroupName).To(Equal("some-security-group"))
					Expect(spaceGUID).To(Equal("some-space-guid"))
					Expect(lifecycle).To(Equal(ccv2.SecurityGroupLifecycle("some-lifecycle")))
				})

				Context("when the actor returns a security group not found error", func() {
					BeforeEach(func() {
						fakeActor.UnbindSecurityGroupByNameAndSpaceReturns(
							v2action.Warnings{"unbind warning"},
							actionerror.SecurityGroupNotFoundError{Name: "some-security-group"},
						)
					})

					It("returns a translated security group not found error", func() {
						Expect(testUI.Err).To(Say("unbind warning"))

						Expect(executeErr).To(MatchError(actionerror.SecurityGroupNotFoundError{Name: "some-security-group"}))
					})
				})

				Context("when the actor returns a security group not bound error", func() {
					BeforeEach(func() {
						fakeActor.UnbindSecurityGroupByNameAndSpaceReturns(
							v2action.Warnings{"unbind warning"},
							actionerror.SecurityGroupNotBoundError{
								Name:      "some-security-group",
								Lifecycle: "some-lifecycle",
							})
					})

					It("returns a translated security group not bound warning but has no error", func() {
						Expect(testUI.Err).To(Say("unbind warning"))
						Expect(testUI.Err).To(Say("Security group some-security-group not bound to this space for lifecycle phase 'some-lifecycle'."))

						Expect(testUI.Out).To(Say("OK"))
						Expect(testUI.Out).NotTo(Say("TIP: Changes require an app restart \\(for running\\) or restage \\(for staging\\) to apply to existing applications\\."))

						Expect(executeErr).NotTo(HaveOccurred())
					})
				})

				Context("when the actor returns an error", func() {
					var expectedErr error

					BeforeEach(func() {
						expectedErr = errors.New("some unbind security group error")
						fakeActor.UnbindSecurityGroupByNameAndSpaceReturns(
							v2action.Warnings{"unbind warning"},
							expectedErr,
						)
					})

					It("returns a translated security no found error", func() {
						Expect(testUI.Err).To(Say("unbind warning"))

						Expect(executeErr).To(MatchError(expectedErr))
					})
				})
			})
		})

		Context("when the security group, org, and space are provided", func() {
			BeforeEach(func() {
				cmd.RequiredArgs.SecurityGroupName = "some-security-group"
				cmd.RequiredArgs.OrganizationName = "some-org"
				cmd.RequiredArgs.SpaceName = "some-space"
			})

			Context("when checking target fails", func() {
				BeforeEach(func() {
					fakeSharedActor.CheckTargetReturns(actionerror.NoOrganizationTargetedError{BinaryName: binaryName})
				})

				It("returns an error", func() {
					Expect(executeErr).To(MatchError(actionerror.NoOrganizationTargetedError{BinaryName: "faceman"}))

					Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
					checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
					Expect(checkTargetedOrg).To(BeFalse())
					Expect(checkTargetedSpace).To(BeFalse())
				})
			})

			Context("when the user is logged in", func() {
				BeforeEach(func() {
					fakeActor.UnbindSecurityGroupByNameOrganizationNameAndSpaceNameReturns(
						v2action.Warnings{"unbind warning"},
						nil)
				})

				It("the security group is unbound from the targeted space", func() {
					Expect(testUI.Out).To(Say("Unbinding security group %s from org %s / space %s as %s\\.\\.\\.", "some-security-group", "some-org", "some-space", "some-user"))
					Expect(testUI.Out).To(Say("OK\n\n"))
					Expect(testUI.Out).To(Say("TIP: Changes require an app restart \\(for running\\) or restage \\(for staging\\) to apply to existing applications\\."))
					Expect(testUI.Err).To(Say("unbind warning"))

					Expect(fakeActor.UnbindSecurityGroupByNameOrganizationNameAndSpaceNameCallCount()).To(Equal(1))
					securityGroupName, orgName, spaceName, lifecycle := fakeActor.UnbindSecurityGroupByNameOrganizationNameAndSpaceNameArgsForCall(0)
					Expect(securityGroupName).To(Equal("some-security-group"))
					Expect(orgName).To(Equal("some-org"))
					Expect(spaceName).To(Equal("some-space"))
					Expect(lifecycle).To(Equal(ccv2.SecurityGroupLifecycle("some-lifecycle")))
				})
			})

			Context("when the actor returns a security group not found error", func() {
				BeforeEach(func() {
					fakeActor.UnbindSecurityGroupByNameOrganizationNameAndSpaceNameReturns(
						v2action.Warnings{"unbind warning"},
						actionerror.SecurityGroupNotFoundError{Name: "some-security-group"},
					)
				})

				It("returns a translated security group not found error", func() {
					Expect(testUI.Err).To(Say("unbind warning"))

					Expect(executeErr).To(MatchError(actionerror.SecurityGroupNotFoundError{Name: "some-security-group"}))
				})
			})

			Context("when the actor returns a security group not bound error", func() {
				BeforeEach(func() {
					fakeActor.UnbindSecurityGroupByNameOrganizationNameAndSpaceNameReturns(
						v2action.Warnings{"unbind warning"},
						actionerror.SecurityGroupNotBoundError{
							Name:      "some-security-group",
							Lifecycle: ccv2.SecurityGroupLifecycle("some-lifecycle"),
						})
				})

				It("returns a translated security group not bound warning but has no error", func() {
					Expect(testUI.Err).To(Say("unbind warning"))
					Expect(testUI.Err).To(Say("Security group some-security-group not bound to this space for lifecycle phase 'some-lifecycle'."))

					Expect(testUI.Out).To(Say("OK"))
					Expect(testUI.Out).NotTo(Say("TIP: Changes require an app restart \\(for running\\) or restage \\(for staging\\) to apply to existing applications\\."))

					Expect(executeErr).NotTo(HaveOccurred())
				})
			})

			Context("when the actor returns an error", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("some unbind security group error")
					fakeActor.UnbindSecurityGroupByNameOrganizationNameAndSpaceNameReturns(
						v2action.Warnings{"unbind warning"},
						expectedErr,
					)
				})

				It("returns a translated security no found error", func() {
					Expect(testUI.Err).To(Say("unbind warning"))

					Expect(executeErr).To(MatchError(expectedErr))
				})
			})
		})

		Context("when the security group and org are provided, but the space is not", func() {
			BeforeEach(func() {
				cmd.RequiredArgs.SecurityGroupName = "some-security-group"
				cmd.RequiredArgs.OrganizationName = "some-org"
			})

			It("an error is returned", func() {
				Expect(executeErr).To(MatchError(translatableerror.ThreeRequiredArgumentsError{
					ArgumentName1: "SECURITY_GROUP",
					ArgumentName2: "ORG",
					ArgumentName3: "SPACE"}))

				Expect(testUI.Out).NotTo(Say("Unbinding security group"))

				Expect(fakeActor.UnbindSecurityGroupByNameOrganizationNameAndSpaceNameCallCount()).To(Equal(0))
			})
		})
	})

	Context("when lifecycle is 'running'", func() {
		BeforeEach(func() {
			cmd.Lifecycle = flag.SecurityGroupLifecycle(ccv2.SecurityGroupLifecycleRunning)
			fakeActor.CloudControllerAPIVersionReturns("2.34.0")
		})

		It("does no version check", func() {
			Expect(executeErr).NotTo(HaveOccurred())
		})
	})

	Context("when lifecycle is 'staging'", func() {
		BeforeEach(func() {
			cmd.Lifecycle = flag.SecurityGroupLifecycle(ccv2.SecurityGroupLifecycleStaging)
		})

		Context("when the version check fails", func() {
			BeforeEach(func() {
				fakeActor.CloudControllerAPIVersionReturns("2.34.0")
			})

			It("returns a MinimumAPIVersionNotMetError", func() {
				Expect(executeErr).To(MatchError(translatableerror.LifecycleMinimumAPIVersionNotMetError{
					CurrentVersion: "2.34.0",
					MinimumVersion: ccversion.MinVersionLifecyleStagingV2,
				}))
				Expect(fakeActor.CloudControllerAPIVersionCallCount()).To(Equal(1))
				Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(0))
			})
		})
	})
})
