package v2_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v2"
	"code.cloudfoundry.org/cli/command/v2/v2fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("security-groups Command", func() {
	var (
		cmd             SecurityGroupsCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v2fakes.FakeSecurityGroupsActor
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v2fakes.FakeSecurityGroupsActor)

		cmd = SecurityGroupsCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)

		fakeConfig.CurrentUserReturns(configv3.User{Name: "some-user"}, nil)
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

		It("returns the error", func() {
			Expect(executeErr).To(MatchError(expectedErr))
		})
	})

	Context("when checking target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(sharedaction.NotLoggedInError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(command.NotLoggedInError{BinaryName: binaryName}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			_, checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeFalse())
			Expect(checkTargetedSpace).To(BeFalse())
		})
	})

	Context("when the list of security groups is returned", func() {
		var secGroups []v2action.SecurityGroupWithOrganizationAndSpace

		BeforeEach(func() {
			secGroups = []v2action.SecurityGroupWithOrganizationAndSpace{
				{
					SecurityGroup: &v2action.SecurityGroup{Name: "seg-group-1"},
					Organization:  &v2action.Organization{Name: "org-11"},
					Space:         &v2action.Space{Name: "space-111"},
				},
				{
					SecurityGroup: &v2action.SecurityGroup{Name: "seg-group-1"},
					Organization:  &v2action.Organization{Name: "org-12"},
					Space:         &v2action.Space{Name: "space-121"},
				},
				{
					SecurityGroup: &v2action.SecurityGroup{Name: "seg-group-1"},
					Organization:  &v2action.Organization{Name: "org-12"},
					Space:         &v2action.Space{Name: "space-122"},
				},
				{
					SecurityGroup: &v2action.SecurityGroup{Name: "seg-group-2"},
					Organization:  &v2action.Organization{},
					Space:         &v2action.Space{},
				},
				{
					SecurityGroup: &v2action.SecurityGroup{Name: "seg-group-3"},
					Organization:  &v2action.Organization{Name: "org-31"},
					Space:         &v2action.Space{Name: "space-311"},
				},
			}
			fakeActor.GetSecurityGroupsWithOrganizationAndSpaceReturns(secGroups, v2action.Warnings{"warning-1", "warning-2"}, nil)
		})

		It("displays a table containing the security groups, the spaces to which they are bound, and the spaces' orgs", func() {
			Expect(executeErr).To(BeNil())

			Expect(fakeActor.GetSecurityGroupsWithOrganizationAndSpaceCallCount()).To(Equal(1))

			Expect(testUI.Out).To(Say("Getting security groups as some-user\\.\\.\\."))
			Expect(testUI.Out).To(Say("OK\\n\\n"))
			Expect(testUI.Out).To(Say("\\s+name\\s+organization\\s+space"))
			Expect(testUI.Out).To(Say("#0\\s+seg-group-1\\s+org-11\\s+space-111"))
			Expect(testUI.Out).To(Say("(?m)\\s+seg-group-1\\s+org-12\\s+space-121"))
			Expect(testUI.Out).To(Say("(?m)\\s+seg-group-1\\s+org-12\\s+space-122"))
			Expect(testUI.Out).To(Say("#1\\s+seg-group-2\\s+"))
			Expect(testUI.Out).To(Say("#2\\s+seg-group-3\\s+org-31\\s+space-311"))
			Expect(testUI.Err).To(Say("warning-1"))
			Expect(testUI.Err).To(Say("warning-2"))
		})
	})

	Context("when an error is encountered fetching the security groups", func() {
		BeforeEach(func() {
			fakeActor.GetSecurityGroupsWithOrganizationAndSpaceReturns(nil, v2action.Warnings{"warning-1", "warning-2"}, errors.New("generic"))
		})

		It("returns the error", func() {
			Expect(executeErr).To(MatchError("generic"))

			Expect(testUI.Err).To(Say("warning-1"))
			Expect(testUI.Err).To(Say("warning-2"))
		})
	})
})
