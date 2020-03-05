package v7_test

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("Security Group Command", func() {
	var (
		cmd             SecurityGroupCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeSecurityGroupActor
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeSecurityGroupActor)
		fakeConfig.TargetedOrganizationNameReturns("some-org")

		cmd = SecurityGroupCommand{
			UI:           testUI,
			Config:       fakeConfig,
			SharedActor:  fakeSharedActor,
			Actor:        fakeActor,
			RequiredArgs: flag.SecurityGroup{SecurityGroup: "some-security-group"},
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

	When("getting the security group fails", func() {
		BeforeEach(func() {
			fakeConfig.CurrentUserReturns(
				configv3.User{
					Name: "some-user",
				},
				nil)

			fakeActor.GetSecurityGroupReturns(
				v7action.SecurityGroupSummary{},
				v7action.Warnings{"warning-1", "warning-2"},
				actionerror.SecurityGroupNotFoundError{})
		})

		It("returns a translatable error and outputs all warnings", func() {
			Expect(testUI.Out).To(Say("Getting info for security group some-security-group as some-user..."))

			Expect(executeErr).To(MatchError(actionerror.SecurityGroupNotFoundError{}))
			Expect(fakeActor.GetSecurityGroupCallCount()).To(Equal(1))
			Expect(testUI.Err).To(Say("warning-1"))
			Expect(testUI.Err).To(Say("warning-2"))
		})
	})

	When("getting the security group succeeds", func() {
		BeforeEach(func() {
			fakeConfig.CurrentUserReturns(configv3.User{Name: "some-user"}, nil)
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "some-org", GUID: "some-org-guid"})
		})

		When("the security group does not have associated rules or spaces", func() {
			BeforeEach(func() {
				fakeActor.GetSecurityGroupReturns(
					v7action.SecurityGroupSummary{
						Name:                "some-security-group",
						Rules:               []resources.Rule{},
						SecurityGroupSpaces: []v7action.SecurityGroupSpace{},
					},
					v7action.Warnings{"warning-1", "warning-2"},
					nil)
			})

			It("displays the security group and all warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeActor.GetSecurityGroupCallCount()).To(Equal(1))
				securityGroupName := fakeActor.GetSecurityGroupArgsForCall(0)
				Expect(securityGroupName).To(Equal("some-security-group"))

				Expect(testUI.Out).To(Say("Getting info for security group some-security-group as some-user..."))
				Expect(testUI.Err).To(Say("warning-1"))
				Expect(testUI.Err).To(Say("warning-2"))

				Expect(testUI.Out).To(Say(`name:\s+some-security-group`))
				Expect(testUI.Out).To(Say(`rules:`))
				Expect(testUI.Out).To(Say(`\[\]`))
				Expect(testUI.Out).To(Say(`No spaces assigned`))
			})
		})

		When("the security group has associated rules and spaces", func() {
			BeforeEach(func() {
				port := "4747"
				description := "Top 8 Friends Only"

				fakeActor.GetSecurityGroupReturns(
					v7action.SecurityGroupSummary{
						Name: "some-security-group",
						Rules: []resources.Rule{{
							Description: &description,
							Destination: "130.58.1.2",
							Ports:       &port,
							Protocol:    "tcp",
						}},
						SecurityGroupSpaces: []v7action.SecurityGroupSpace{{
							OrgName:   "obsolete-social-networks",
							SpaceName: "my-space",
						}},
					},
					v7action.Warnings{"warning-1", "warning-2"},
					nil)
			})

			It("displays the security group and all warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeActor.GetSecurityGroupCallCount()).To(Equal(1))
				securityGroupName := fakeActor.GetSecurityGroupArgsForCall(0)
				Expect(securityGroupName).To(Equal("some-security-group"))

				Expect(testUI.Out).To(Say("Getting info for security group some-security-group as some-user..."))
				Expect(testUI.Err).To(Say("warning-1"))
				Expect(testUI.Err).To(Say("warning-2"))

				Expect(testUI.Out).To(Say(`name:\s+some-security-group`))
				Expect(testUI.Out).To(Say(`rules:`))
				Expect(testUI.Out).To(Say(`\[`))
				Expect(testUI.Out).To(Say(`{`))
				Expect(testUI.Out).To(Say(`\s+"protocol":\s+"tcp"`))
				Expect(testUI.Out).To(Say(`\s+"destination":\s+"130.58.1.2"`))
				Expect(testUI.Out).To(Say(`\s+"ports":\s+"4747"`))
				Expect(testUI.Out).To(Say(`\s+"description":\s+"Top 8 Friends Only"`))
				Expect(testUI.Out).To(Say(`}`))
				Expect(testUI.Out).To(Say(`\]`))
				Expect(testUI.Out).To(Say(`organization\s+space`))
				Expect(testUI.Out).To(Say(`obsolete-social-networks\s+my-space`))
			})
		})
	})
})
