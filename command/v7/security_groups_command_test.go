package v7_test

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/cf/errors"
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

var _ = Describe("Security Groups Command", func() {
	var (
		cmd             SecurityGroupsCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)
		fakeConfig.TargetedOrganizationNameReturns("some-org")

		cmd = SecurityGroupsCommand{
			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
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

	When("getting the security groups fails", func() {
		BeforeEach(func() {
			fakeConfig.CurrentUserReturns(
				configv3.User{
					Name: "some-user",
				},
				nil)

			fakeActor.GetSecurityGroupsReturns(
				[]v7action.SecurityGroupSummary{},
				v7action.Warnings{"warning-1", "warning-2"},
				errors.New("Aah I'm an error"))
		})

		It("returns a translatable error and outputs all warnings", func() {
			Expect(testUI.Out).To(Say("Getting security groups as some-user..."))

			Expect(executeErr).To(MatchError(errors.New("Aah I'm an error")))
			Expect(fakeActor.GetSecurityGroupsCallCount()).To(Equal(1))
			Expect(testUI.Err).To(Say("warning-1"))
			Expect(testUI.Err).To(Say("warning-2"))
		})
	})

	When("getting the security groups succeeds", func() {
		BeforeEach(func() {
			fakeConfig.CurrentUserReturns(configv3.User{Name: "some-user"}, nil)
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "some-org", GUID: "some-org-guid"})
		})

		When("there are no security groups", func() {
			BeforeEach(func() {
				fakeActor.GetSecurityGroupsReturns(
					[]v7action.SecurityGroupSummary{},
					v7action.Warnings{"warning-1", "warning-2"},
					nil,
				)
			})

			It("displays an empty state message", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeActor.GetSecurityGroupsCallCount()).To(Equal(1))

				Expect(testUI.Out).To(Say("Getting security groups as some-user..."))
				Expect(testUI.Err).To(Say("warning-1"))
				Expect(testUI.Err).To(Say("warning-2"))

				Expect(testUI.Out).To(Say("No security groups found."))
			})
		})

		When("there are security groups", func() {
			BeforeEach(func() {
				fakeActor.GetSecurityGroupsReturns(
					[]v7action.SecurityGroupSummary{{
						Name:  "security-group-1",
						Rules: []resources.Rule{},
						SecurityGroupSpaces: []v7action.SecurityGroupSpace{{
							OrgName:   "org-1",
							SpaceName: "space-1",
							Lifecycle: "running",
						}, {
							OrgName:   "<all>",
							SpaceName: "<all>",
							Lifecycle: "staging",
						}, {
							OrgName:   "org-1",
							SpaceName: "space-1",
							Lifecycle: "staging",
						}},
					}, {
						Name:  "security-group-2",
						Rules: []resources.Rule{},
						SecurityGroupSpaces: []v7action.SecurityGroupSpace{{
							OrgName:   "<all>",
							SpaceName: "<all>",
							Lifecycle: "running",
						}},
					}, {
						Name:                "security-group-3",
						Rules:               []resources.Rule{},
						SecurityGroupSpaces: []v7action.SecurityGroupSpace{},
					}},
					v7action.Warnings{"warning-1", "warning-2"},
					nil,
				)
			})

			It("displays the security groups", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeActor.GetSecurityGroupsCallCount()).To(Equal(1))

				Expect(testUI.Out).To(Say("Getting security groups as some-user..."))
				Expect(testUI.Err).To(Say("warning-1"))
				Expect(testUI.Err).To(Say("warning-2"))

				Expect(testUI.Out).To(Say(`name\s+organization\s+space\s+lifecycle`))
				Expect(testUI.Out).To(Say(`security-group-1\s+org-1\s+space-1\s+running`))
				Expect(testUI.Out).To(Say(`security-group-1\s+<all>\s+<all>\s+staging`))
				Expect(testUI.Out).To(Say(`security-group-1\s+org-1\s+space-1\s+staging`))
				Expect(testUI.Out).To(Say(`security-group-2\s+<all>\s+<all>\s+running`))
				Expect(testUI.Out).To(Say(`security-group-3\s+`))
			})
		})
	})
})
