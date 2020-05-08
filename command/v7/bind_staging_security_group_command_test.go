package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("bind-staging-security-group Command", func() {
	var (
		cmd               BindStagingSecurityGroupCommand
		testUI            *ui.UI
		fakeConfig        *commandfakes.FakeConfig
		fakeSharedActor   *commandfakes.FakeSharedActor
		fakeActor         *v7fakes.FakeActor
		binaryName        string
		executeErr        error
		securityGroupName = "sec-group-name"
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		cmd = BindStagingSecurityGroupCommand{
			BaseCommand: command.BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
			SecurityGroup: flag.SecurityGroup{SecurityGroup: securityGroupName},
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)

		fakeConfig.CurrentUserReturns(
			configv3.User{Name: "some-user"},
			nil)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})
	BeforeEach(func() {
		fakeActor.UpdateSecurityGroupGloballyEnabledReturns(
			v7action.Warnings{"globally bind security group warning"},
			nil)
	})

	When("the current user is invalid", func() {
		BeforeEach(func() {
			fakeConfig.CurrentUserReturns(
				configv3.User{},
				errors.New("some-error"))
		})
		It("returns the error", func() {
			Expect(executeErr).To(HaveOccurred())
			Expect(executeErr).To(MatchError("some-error"))
		})
	})

	It("globally binds the security group displays all warnings", func() {
		Expect(testUI.Out).To(Say(`Binding security group sec-group-name to staging as some-user\.\.\.`))

		Expect(fakeActor.UpdateSecurityGroupGloballyEnabledCallCount()).To(Equal(1))
		securityGroupName, lifecycle, enabled := fakeActor.UpdateSecurityGroupGloballyEnabledArgsForCall(0)
		Expect(securityGroupName).To(Equal("sec-group-name"))
		Expect(lifecycle).To(Equal(constant.SecurityGroupLifecycleStaging))
		Expect(enabled).To(Equal(true))
	})

	When("an error is encountered globally binding the security group", func() {
		var expectedErr error

		BeforeEach(func() {
			expectedErr = errors.New("bind security group error")
			fakeActor.UpdateSecurityGroupGloballyEnabledReturns(
				v7action.Warnings{"globally bind security group warning"},
				expectedErr)
		})

		It("returns the error and displays all warnings", func() {
			Expect(executeErr).To(MatchError(expectedErr))

			Expect(testUI.Out).NotTo(Say("OK"))

			Expect(testUI.Err).To(Say("globally bind security group warning"))
		})
	})

	It("signals that the security group is bound and provides a helpful tip", func() {
		Expect(testUI.Out).To(Say("OK"))
		Expect(testUI.Out).To(Say(`TIP: Changes require an app restart \(for running\) or restage \(for staging\) to apply to existing applications\.`))
		Expect(testUI.Err).To(Say("globally bind security group warning"))
		Expect(executeErr).NotTo(HaveOccurred())
	})
})
