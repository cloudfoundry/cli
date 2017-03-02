package v2_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/v2"
	"code.cloudfoundry.org/cli/command/v2/shared"
	"code.cloudfoundry.org/cli/command/v2/v2fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("bind-security-group Command", func() {
	var (
		cmd             v2.BindSecurityGroupCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v2fakes.FakeBindSecurityGroupActor
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v2fakes.FakeBindSecurityGroupActor)

		cmd = v2.BindSecurityGroupCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

		cmd.RequiredArgs.SecurityGroupName = "some-security-group"
		cmd.RequiredArgs.OrganizationName = "some-org"

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		// TODO: remove when experimental flag is removed
		fakeConfig.ExperimentalReturns(true)

		// Stubs for the happy path
		fakeConfig.CurrentUserReturns(
			configv3.User{Name: "some-user"},
			nil)
		fakeActor.GetSecurityGroupByNameReturns(
			v2action.SecurityGroup{Name: "some-security-group", GUID: "some-security-group-guid"},
			v2action.Warnings{"get security group warning"},
			nil,
		)
		fakeActor.GetOrganizationByNameReturns(
			v2action.Organization{Name: "some-org", GUID: "some-org-guid"},
			v2action.Warnings{"get org warning"},
			nil,
		)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	// TODO: remove when experimental flag is removed
	It("Displays the experimental warning message", func() {
		Expect(testUI.Out).To(Say(command.ExperimentalWarning))
	})

	Context("when checking target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(sharedaction.NotLoggedInError{BinaryName: binaryName})
		})

		It("returns an error if the check fails", func() {
			Expect(executeErr).To(MatchError(command.NotLoggedInError{BinaryName: "faceman"}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			_, checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeFalse())
			Expect(checkTargetedSpace).To(BeFalse())
		})
	})

	Context("when getting the current user returns an error", func() {
		var expectedErr error

		BeforeEach(func() {
			expectedErr = errors.New("getting current user error")
			fakeConfig.CurrentUserReturns(
				configv3.User{},
				expectedErr)
		})

		It("returns the error", func() {
			Expect(executeErr).To(MatchError(expectedErr))
		})
	})

	Context("when the provided security group does not exist", func() {
		BeforeEach(func() {
			fakeActor.GetSecurityGroupByNameReturns(
				v2action.SecurityGroup{},
				v2action.Warnings{"get security group warning"},
				v2action.SecurityGroupNotFoundError{Name: "some-security-group"},
			)
		})

		It("returns a SecurityGroupNotFoundError and displays all warnings", func() {
			Expect(executeErr).To(MatchError(shared.SecurityGroupNotFoundError{Name: "some-security-group"}))
			Expect(testUI.Err).To(Say("get security group warning"))
		})
	})

	Context("when the provided org does not exist", func() {
		BeforeEach(func() {
			fakeActor.GetOrganizationByNameReturns(
				v2action.Organization{},
				v2action.Warnings{"get organization warning"},
				v2action.OrganizationNotFoundError{Name: "some-org"},
			)
		})

		It("returns an OrganizationNotFoundError and displays all warnings", func() {
			Expect(executeErr).To(MatchError(shared.OrganizationNotFoundError{Name: "some-org"}))
			Expect(testUI.Err).To(Say("get organization warning"))
		})
	})

	Context("when a space is provided", func() {
		BeforeEach(func() {
			cmd.RequiredArgs.SpaceName = "some-space"
		})

		Context("when the space does not exist", func() {
			BeforeEach(func() {
				fakeActor.GetSpaceByOrganizationAndNameReturns(
					v2action.Space{},
					v2action.Warnings{"get space warning"},
					v2action.SpaceNotFoundError{Name: "some-space"},
				)
			})

			It("returns a SpaceNotFoundError", func() {
				Expect(executeErr).To(MatchError(shared.SpaceNotFoundError{Name: "some-space"}))
				Expect(testUI.Err).To(Say("get space warning"))
			})
		})

		Context("when the space does exist", func() {
			BeforeEach(func() {
				fakeActor.GetSpaceByOrganizationAndNameReturns(
					v2action.Space{GUID: "some-space-guid"},
					v2action.Warnings{"get space by org warning"},
					nil,
				)
			})

			It("binds the security group to the space and displays all warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(fakeActor.BindSecurityGroupToSpaceCallCount()).To(Equal(1))
				securityGroupGUID, spaceGUID := fakeActor.BindSecurityGroupToSpaceArgsForCall(0)
				Expect(securityGroupGUID).To(Equal("some-security-group-guid"))
				Expect(spaceGUID).To(Equal("some-space-guid"))

				Expect(testUI.Err).To(Say("get security group warning"))
				Expect(testUI.Err).To(Say("get org warning"))
				Expect(testUI.Err).To(Say("get space by org warning"))
			})
		})
	})
})
