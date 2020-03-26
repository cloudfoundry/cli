package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
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
		fakeActor       *v7fakes.FakeActor
		binaryName      string
		executeErr      error
		expectedErr     error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		cmd = UnbindSecurityGroupCommand{
			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		cmd.RequiredArgs.SecurityGroupName = "some-security-group"
		cmd.RequiredArgs.OrganizationName = "some-org"
		cmd.RequiredArgs.SpaceName = "some-space"
		cmd.Lifecycle = "some-lifecycle"

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		fakeConfig.ExperimentalReturns(true)

		fakeConfig.CurrentUserReturns(
			configv3.User{Name: "some-user"},
			nil)
		fakeActor.UnbindSecurityGroupReturns(
			v7action.Warnings{"unbind warning"},
			nil)
		fakeSharedActor.CheckTargetReturns(nil)

	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("checking target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: "faceman"}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeFalse())
			Expect(checkTargetedSpace).To(BeFalse())
		})
	})

	When("getting the current user fails", func() {
		BeforeEach(func() {
			expectedErr = errors.New("getting user failed")
			fakeConfig.CurrentUserReturns(configv3.User{}, expectedErr)
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(expectedErr))
		})
	})

	It("indicates that it will unbind the security group", func() {
		Expect(testUI.Out).To(Say(`Unbinding security group %s from org %s / space %s as %s\.\.\.`, "some-security-group", "some-org", "some-space", "some-user"))
	})

	It("displays the warnings from unbinding the security group", func() {
		Expect(testUI.Err).To(Say("unbind warning"))
	})

	When("unbinding the security group fails", func() {
		BeforeEach(func() {
			fakeActor.UnbindSecurityGroupReturns(
				v7action.Warnings{"unbind warning"},
				errors.New("unbind-error"),
			)
		})
		When("security group is unbound", func() {
			BeforeEach(func() {
				fakeActor.UnbindSecurityGroupReturns(
					v7action.Warnings{"unbind warning"},
					actionerror.SecurityGroupNotBoundToSpaceError{
						Name:      "some-security-group",
						Space:     "some-space",
						Lifecycle: constant.SecurityGroupLifecycle("some-lifecycle"),
					})
			})
			It("returns the security group not bound error as a warning and succeeds", func() {
				Expect(fakeActor.UnbindSecurityGroupCallCount()).To(Equal(1))
				securityGroupName, orgName, spaceName, lifecycle := fakeActor.UnbindSecurityGroupArgsForCall(0)
				Expect(securityGroupName).To(Equal("some-security-group"))
				Expect(orgName).To(Equal("some-org"))
				Expect(spaceName).To(Equal("some-space"))
				Expect(lifecycle).To(Equal(constant.SecurityGroupLifecycle("some-lifecycle")))
				Expect(testUI.Err).To(Say("Security group some-security-group not bound to space some-space for lifecycle phase 'some-lifecycle'."))
				Expect(testUI.Out).To(Say("OK\n\n"))
				Expect(testUI.Out).To(Say(`TIP: Changes require an app restart \(for running\) or restage \(for staging\) to apply to existing applications\.`))
				Expect(executeErr).NotTo(HaveOccurred())
			})
		})
		It("returns the error", func() {
			Expect(fakeActor.UnbindSecurityGroupCallCount()).To(Equal(1))
			securityGroupName, orgName, spaceName, lifecycle := fakeActor.UnbindSecurityGroupArgsForCall(0)
			Expect(securityGroupName).To(Equal("some-security-group"))
			Expect(orgName).To(Equal("some-org"))
			Expect(spaceName).To(Equal("some-space"))
			Expect(lifecycle).To(Equal(constant.SecurityGroupLifecycle("some-lifecycle")))
			Expect(executeErr).To(MatchError("unbind-error"))
		})
	})

	It("indicates it successfully unbound the security group", func() {
		Expect(testUI.Out).To(Say("OK\n\n"))
		Expect(testUI.Out).To(Say(`TIP: Changes require an app restart \(for running\) or restage \(for staging\) to apply to existing applications\.`))
	})

})
