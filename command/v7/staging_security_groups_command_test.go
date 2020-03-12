package v7_test

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("Staging Security Groups Command", func() {
	var (
		cmd             StagingSecurityGroupsCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeStagingSecurityGroupsActor
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeStagingSecurityGroupsActor)
		fakeConfig.TargetedOrganizationNameReturns("some-org")

		cmd = StagingSecurityGroupsCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("checking the target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(
				actionerror.NotLoggedInError{BinaryName: "binaryName"})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: "binaryName"}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			targetedOrganizationRequired, targetedSpaceRequired := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(targetedOrganizationRequired).To(Equal(false))
			Expect(targetedSpaceRequired).To(Equal(false))
		})
	})

	When("there are no globally enabled staging security groups found", func() {
		BeforeEach(func() {
			fakeConfig.CurrentUserReturns(
				configv3.User{
					Name: "some-user",
				},
				nil)

			fakeActor.GetGlobalStagingSecurityGroupsReturns(
				[]resources.SecurityGroup{},
				v7action.Warnings{"warning-1", "warning-2"},
				nil,
			)
		})

		It("returns a translatable error and outputs all warnings", func() {
			Expect(testUI.Out).To(Say("Getting global staging security groups as some-user..."))
			Expect(testUI.Out).To(Say("No global staging security groups found."))

			Expect(fakeActor.GetGlobalStagingSecurityGroupsCallCount()).To(Equal(1))
			Expect(testUI.Err).To(Say("warning-1"))
			Expect(testUI.Err).To(Say("warning-2"))
		})
	})

	When("there are globally enabled staging security groups", func() {
		BeforeEach(func() {
			fakeConfig.CurrentUserReturns(configv3.User{Name: "some-user"}, nil)
		})

		When("the security group does not have associated rules or spaces", func() {
			BeforeEach(func() {
				fakeActor.GetGlobalStagingSecurityGroupsReturns(
					[]resources.SecurityGroup{{
						Name: "staging-security-group-1",
					}, {
						Name: "staging-security-group-2",
					}},
					v7action.Warnings{"warning-1", "warning-2"},
					nil)
			})

			It("displays the security groups and all warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeActor.GetGlobalStagingSecurityGroupsCallCount()).To(Equal(1))

				Expect(testUI.Out).To(Say("Getting global staging security groups as some-user..."))
				Expect(testUI.Err).To(Say("warning-1"))
				Expect(testUI.Err).To(Say("warning-2"))

				Expect(testUI.Out).To(Say(`name`))
				Expect(testUI.Out).To(Say(`staging-security-group-1`))
				Expect(testUI.Out).To(Say(`staging-security-group-2`))
			})
		})
	})
})
