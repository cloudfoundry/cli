package v7_test

import (
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"

	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("set-space-role Command", func() {
	var (
		cmd             SetSpaceRoleCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeSetSpaceRoleActor
		binaryName      string
		executeErr      error
		input           *Buffer
	)

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeSetSpaceRoleActor)

		cmd = SetSpaceRoleCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	BeforeEach(func() {
		fakeConfig.CurrentUserReturns(configv3.User{Name: "current-user"}, nil)

		fakeActor.GetOrganizationByNameReturns(
			v7action.Organization{
				GUID:     "some-org-guid",
				Name:     "some-org-name",
				Metadata: nil,
			},
			v7action.Warnings{"get-org-warning"},
			nil,
		)

		fakeActor.GetSpaceByNameAndOrganizationReturns(
			v7action.Space{GUID: "some-space-guid", Name: "some-space-name"},
			v7action.Warnings{"get-space-warning"},
			nil,
		)
	})

	When("neither origin nor client flag is provided", func() {
		BeforeEach(func() {
			cmd.Args.Organization = "some-org-name"
			cmd.Args.Space = "some-space-name"
			cmd.Args.Role = flag.SpaceRole{Role: "SpaceDeveloper"}
			cmd.Args.Username = "target-user-name"
		})

		It("creates the org user role by username in the default origin", func() {
			Expect(fakeActor.CreateOrgRoleByUserNameCallCount()).To(Equal(1))
			givenRoleType, givenUserName, givenOrigin, givenOrgGUID := fakeActor.CreateOrgRoleByUserNameArgsForCall(0)
			Expect(givenRoleType).To(Equal(constant.OrgUserRole))
			Expect(givenUserName).To(Equal("target-user-name"))
			Expect(givenOrigin).To(Equal("uaa"))
			Expect(givenOrgGUID).To(Equal("some-org-guid"))
		})

		It("creates the space role", func() {
			Expect(fakeActor.CreateSpaceRoleByUserNameCallCount()).To(Equal(1))
			givenRoleType, givenUserName, givenOrigin, givenSpaceGUID := fakeActor.CreateSpaceRoleByUserNameArgsForCall(0)
			Expect(givenRoleType).To(Equal(constant.SpaceDeveloperRole))
			Expect(givenUserName).To(Equal("target-user-name"))
			Expect(givenOrigin).To(Equal("uaa"))
			Expect(givenSpaceGUID).To(Equal("some-space-guid"))
		})

		It("displays flavor text and returns without error", func() {
			Expect(testUI.Out).To(Say("Assigning role SpaceDeveloper to user target-user-name in org some-org-name / space some-space-name as current-user..."))
			Expect(testUI.Out).To(Say("OK"))
			Expect(executeErr).NotTo(HaveOccurred())
		})
	})

	When("origin flag is provided", func() {
		BeforeEach(func() {
			cmd.Args.Organization = "some-org-name"
			cmd.Args.Space = "some-space-name"
			cmd.Args.Role = flag.SpaceRole{Role: "SpaceAuditor"}
			cmd.Args.Username = "target-user-name"
			cmd.Origin = "ldap"
		})

		It("creates the space role", func() {
			Expect(fakeActor.CreateSpaceRoleByUserNameCallCount()).To(Equal(1))
			givenRoleType, givenUserName, givenOrigin, givenSpaceGUID := fakeActor.CreateSpaceRoleByUserNameArgsForCall(0)
			Expect(givenRoleType).To(Equal(constant.SpaceAuditorRole))
			Expect(givenUserName).To(Equal("target-user-name"))
			Expect(givenOrigin).To(Equal("ldap"))
			Expect(givenSpaceGUID).To(Equal("some-space-guid"))
		})

		It("displays flavor text and returns without error", func() {
			Expect(testUI.Out).To(Say("Assigning role SpaceAuditor to user target-user-name in org some-org-name / space some-space-name as current-user..."))
			Expect(testUI.Out).To(Say("OK"))
			Expect(executeErr).NotTo(HaveOccurred())
		})
	})

	When("client flag is provided", func() {
		BeforeEach(func() {
			cmd.Args.Organization = "some-org-name"
			cmd.Args.Space = "some-space-name"
			cmd.Args.Role = flag.SpaceRole{Role: "SpaceAuditor"}
			cmd.Args.Username = "target-user-name"
			cmd.IsClient = true
		})

		It("does not try to get the user", func() {
			Expect(fakeActor.GetUserCallCount()).To(Equal(0))
		})

		It("creates the space role by user guid", func() {
			Expect(fakeActor.CreateSpaceRoleByUserGUIDCallCount()).To(Equal(1))
			givenRoleType, givenUserGUID, givenSpaceGUID := fakeActor.CreateSpaceRoleByUserGUIDArgsForCall(0)
			Expect(givenRoleType).To(Equal(constant.SpaceAuditorRole))
			Expect(givenUserGUID).To(Equal("target-user-name"))
			Expect(givenSpaceGUID).To(Equal("some-space-guid"))
		})

		It("displays flavor text and returns without error", func() {
			Expect(testUI.Out).To(Say("Assigning role SpaceAuditor to user target-user-name in org some-org-name / space some-space-name as current-user..."))
			Expect(testUI.Out).To(Say("OK"))
			Expect(executeErr).NotTo(HaveOccurred())
		})
	})

	When("both client and origin flags are provided", func() {
		BeforeEach(func() {
			cmd.Args.Organization = "some-org-name"
			cmd.Args.Space = "some-space-name"
			cmd.Args.Role = flag.SpaceRole{Role: "SpaceAuditor"}
			cmd.Args.Username = "target-user-name"
			cmd.Origin = "ldap"
			cmd.IsClient = true
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(translatableerror.ArgumentCombinationError{
				Args: []string{"--client", "--origin"},
			}))
		})
	})

	When("the role already exists", func() {
		BeforeEach(func() {
			cmd.Args.Organization = "some-org-name"
			cmd.Args.Space = "some-space-name"
			cmd.Args.Role = flag.SpaceRole{Role: "SpaceAuditor"}
			cmd.Args.Username = "target-user-name"

			fakeActor.CreateSpaceRoleByUserNameReturns(
				v7action.Role{},
				v7action.Warnings{"create-role-warning"},
				ccerror.RoleAlreadyExistsError{},
			)
		})

		It("displays warnings and returns without error", func() {
			Expect(testUI.Err).To(Say("create-role-warning"))
			Expect(testUI.Err).To(Say("User 'target-user-name' already has role 'SpaceAuditor' in org 'some-org-name' / space 'some-space-name'."))
			Expect(testUI.Out).To(Say("OK"))
			Expect(executeErr).NotTo(HaveOccurred())
		})
	})

	When("invalid role arg is given", func() {
		BeforeEach(func() {
			cmd.Args.Role = flag.SpaceRole{Role: "Astronaut"}
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError("Invalid role type."))
		})
	})

	When("getting the org fails", func() {
		BeforeEach(func() {
			cmd.Args.Role = flag.SpaceRole{Role: "SpaceAuditor"}

			fakeActor.GetOrganizationByNameReturns(
				v7action.Organization{},
				v7action.Warnings{"get-user-warning"},
				errors.New("get-org-error"),
			)
		})

		It("displays warnings and returns an error", func() {
			Expect(testUI.Err).To(Say("get-user-warning"))
			Expect(executeErr).To(MatchError("get-org-error"))
		})
	})

	When("getting the space fails", func() {
		BeforeEach(func() {
			cmd.Args.Role = flag.SpaceRole{Role: "SpaceAuditor"}

			fakeActor.GetSpaceByNameAndOrganizationReturns(
				v7action.Space{},
				v7action.Warnings{"get-space-warning"},
				errors.New("get-space-error"),
			)
		})

		It("displays warnings and returns an error", func() {
			Expect(testUI.Err).To(Say("get-space-warning"))
			Expect(executeErr).To(MatchError("get-space-error"))
		})
	})

	When("creating the org role fails", func() {
		BeforeEach(func() {
			cmd.Args.Organization = "some-org-name"
			cmd.Args.Space = "some-space-name"
			cmd.Args.Role = flag.SpaceRole{Role: "SpaceAuditor"}
			cmd.Args.Username = "target-user-name"

			fakeActor.CreateOrgRoleByUserNameReturns(
				v7action.Role{},
				v7action.Warnings{"create-org-role-warning"},
				errors.New("create-org-role-error"),
			)
		})

		It("displays warnings and returns without error", func() {
			Expect(testUI.Err).To(Say("create-org-role-warning"))
			Expect(executeErr).To(MatchError("create-org-role-error"))
		})
	})

	When("creating the space role fails", func() {
		BeforeEach(func() {
			cmd.Args.Organization = "some-org-name"
			cmd.Args.Space = "some-space-name"
			cmd.Args.Role = flag.SpaceRole{Role: "SpaceAuditor"}
			cmd.Args.Username = "target-user-name"

			fakeActor.CreateSpaceRoleByUserNameReturns(
				v7action.Role{},
				v7action.Warnings{"create-role-warning"},
				errors.New("create-role-error"),
			)
		})

		It("displays warnings and returns without error", func() {
			Expect(testUI.Err).To(Say("create-role-warning"))
			Expect(executeErr).To(MatchError("create-role-error"))
		})
	})
})
