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

	When("enviroment is set up correctly and all arguments are valid", func() {
		BeforeEach(func() {
			cmd.RequiredArgs.SecurityGroupName = "some-security-group"
			cmd.RequiredArgs.OrganizationName = "some-org"
			cmd.RequiredArgs.SpaceName = "some-space"
			cmd.Lifecycle = "some-lifecycle"
			fakeActor.UnbindSecurityGroupReturns(
				v7action.Warnings{"unbind warning"},
				nil)
		})

		It("the security group is unbound from the targeted space", func() {
			Expect(testUI.Out).To(Say(`Unbinding security group %s from org %s / space %s as %s\.\.\.`, "some-security-group", "some-org", "some-space", "some-user"))
			Expect(testUI.Out).To(Say("OK\n\n"))
			Expect(testUI.Out).To(Say(`TIP: Changes require an app restart \(for running\) or restage \(for staging\) to apply to existing applications\.`))
			Expect(testUI.Err).To(Say("unbind warning"))

			Expect(fakeActor.UnbindSecurityGroupCallCount()).To(Equal(1))
			securityGroupName, orgName, spaceName, lifecycle := fakeActor.UnbindSecurityGroupArgsForCall(0)
			Expect(securityGroupName).To(Equal("some-security-group"))
			Expect(orgName).To(Equal("some-org"))
			Expect(spaceName).To(Equal("some-space"))
			Expect(lifecycle).To(Equal(constant.SecurityGroupLifecycle("some-lifecycle")))
		})
	})

	When("getting the current user fails", func() {
		var expectedErr error

		BeforeEach(func() {
			expectedErr = errors.New("getting user failed")
			fakeConfig.CurrentUserReturns(configv3.User{}, expectedErr)
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(expectedErr))
		})
	})

	When("the actor returns a security group not bound error", func() {
		BeforeEach(func() {
			fakeActor.UnbindSecurityGroupReturns(
				v7action.Warnings{"unbind warning"},
				actionerror.SecurityGroupNotBoundToSpaceError{
					Name:      "some-security-group",
					Lifecycle: constant.SecurityGroupLifecycle("staging"),
				})
		})

		It("returns a translated security group not bound warning but has no error", func() {
			Expect(testUI.Err).To(Say("unbind warning"))
			Expect(testUI.Err).To(Say("Security group 'some-security-group' not bound to this space for the staging lifecycle."))

			Expect(testUI.Out).To(Say("OK"))
			Expect(testUI.Out).NotTo(Say(`TIP: Changes require an app restart \(for running\) or restage \(for staging\) to apply to existing applications\.`))

			Expect(executeErr).NotTo(HaveOccurred())
		})
	})

	When("the actor returns a security group not found error", func() {
		BeforeEach(func() {
			fakeActor.UnbindSecurityGroupReturns(
				v7action.Warnings{"unbind warning"},
				actionerror.SecurityGroupNotFoundError{Name: "some-security-group"},
			)
		})

		It("returns a translated security group not found error", func() {
			Expect(testUI.Err).To(Say("unbind warning"))

			Expect(executeErr).To(MatchError(actionerror.SecurityGroupNotFoundError{Name: "some-security-group"}))
		})
	})

	When("the actor returns an error", func() {
		var expectedErr error

		BeforeEach(func() {
			expectedErr = errors.New("some unbind security group error")
			fakeActor.UnbindSecurityGroupReturns(
				v7action.Warnings{"unbind warning"},
				expectedErr,
			)
		})

		It("returns a translated security no found error", func() {
			Expect(testUI.Err).To(Say("unbind warning"))

			Expect(executeErr).To(MatchError(expectedErr))
		})
	})
})

