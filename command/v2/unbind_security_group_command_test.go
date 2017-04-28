package v2_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v2"
	"code.cloudfoundry.org/cli/command/v2/shared"
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

	Context("when only the security group is provided", func() {
		BeforeEach(func() {
			cmd.RequiredArgs.SecurityGroupName = "some-security-group"
		})

		Context("when checking target fails", func() {
			BeforeEach(func() {
				fakeSharedActor.CheckTargetReturns(sharedaction.NoTargetedOrganizationError{BinaryName: binaryName})
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError(command.NoTargetedOrganizationError{BinaryName: "faceman"}))

				Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
				_, checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
				Expect(checkTargetedOrg).To(BeTrue())
				Expect(checkTargetedSpace).To(BeTrue())
			})
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
				Expect(testUI.Out).To(Say("TIP: Changes will not apply to existing running applications until they are restarted\\."))
				Expect(testUI.Err).To(Say("unbind warning"))

				Expect(fakeConfig.TargetedOrganizationCallCount()).To(Equal(1))
				Expect(fakeConfig.TargetedSpaceCallCount()).To(Equal(1))
				Expect(fakeActor.UnbindSecurityGroupByNameAndSpaceCallCount()).To(Equal(1))
				securityGroupName, spaceGUID := fakeActor.UnbindSecurityGroupByNameAndSpaceArgsForCall(0)
				Expect(securityGroupName).To(Equal("some-security-group"))
				Expect(spaceGUID).To(Equal("some-space-guid"))
			})
		})

		Context("when the actor returns errors", func() {
			BeforeEach(func() {
				cmd.RequiredArgs.SecurityGroupName = "some-security-group"
				fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "some-org"})
				fakeConfig.TargetedSpaceReturns(configv3.Space{Name: "some-space", GUID: "some-space-guid"})
				fakeActor.UnbindSecurityGroupByNameAndSpaceReturns(
					v2action.Warnings{"unbind warning"},
					v2action.SecurityGroupNotFoundError{Name: "some-security-group"},
				)
			})

			It("handles the error", func() {
				Expect(testUI.Err).To(Say("unbind warning"))

				Expect(executeErr).To(MatchError(shared.SecurityGroupNotFoundError{Name: "some-security-group"}))
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
				fakeSharedActor.CheckTargetReturns(sharedaction.NoTargetedOrganizationError{BinaryName: binaryName})
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError(command.NoTargetedOrganizationError{BinaryName: "faceman"}))

				Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
				_, checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
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
				Expect(testUI.Out).To(Say("TIP: Changes will not apply to existing running applications until they are restarted\\."))
				Expect(testUI.Err).To(Say("unbind warning"))

				Expect(fakeActor.UnbindSecurityGroupByNameOrganizationNameAndSpaceNameCallCount()).To(Equal(1))
				securityGroupName, orgName, spaceName := fakeActor.UnbindSecurityGroupByNameOrganizationNameAndSpaceNameArgsForCall(0)
				Expect(securityGroupName).To(Equal("some-security-group"))
				Expect(orgName).To(Equal("some-org"))
				Expect(spaceName).To(Equal("some-space"))
			})
		})

		Context("when the actor returns errors", func() {
			BeforeEach(func() {
				cmd.RequiredArgs.SecurityGroupName = "some-security-group"
				cmd.RequiredArgs.OrganizationName = "some-org"
				cmd.RequiredArgs.SpaceName = "some-space"
				fakeActor.UnbindSecurityGroupByNameOrganizationNameAndSpaceNameReturns(
					v2action.Warnings{"unbind warning"},
					v2action.SecurityGroupNotFoundError{Name: "some-security-group"},
				)
			})

			It("handles the error", func() {
				Expect(testUI.Err).To(Say("unbind warning"))

				Expect(executeErr).To(MatchError(shared.SecurityGroupNotFoundError{Name: "some-security-group"}))
			})
		})
	})

	Context("when the security group and org are provided, but the space is not", func() {
		BeforeEach(func() {
			cmd.RequiredArgs.SecurityGroupName = "some-security-group"
			cmd.RequiredArgs.OrganizationName = "some-org"
		})

		It("an error is returned", func() {
			Expect(executeErr).To(MatchError(command.ThreeRequiredArgumentsError{
				ArgumentName1: "SECURITY_GROUP",
				ArgumentName2: "ORG",
				ArgumentName3: "SPACE"}))

			Consistently(testUI.Out).ShouldNot(Say("Unbinding security group"))

			Expect(fakeActor.UnbindSecurityGroupByNameOrganizationNameAndSpaceNameCallCount()).To(Equal(0))
		})
	})
})
