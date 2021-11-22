package v7_test

import (
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/util/configv3"

	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("set-org-role Command", func() {
	var (
		cmd             SetOrgRoleCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		binaryName      string
		executeErr      error
		input           *Buffer
	)

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		cmd = SetOrgRoleCommand{
			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	BeforeEach(func() {
		fakeActor.GetCurrentUserReturns(configv3.User{Name: "current-user"}, nil)

		fakeActor.GetOrganizationByNameReturns(
			resources.Organization{GUID: "some-org-guid", Name: "some-org-name"},
			v7action.Warnings{"get-org-warning"},
			nil,
		)

		fakeActor.GetUserReturns(
			resources.User{GUID: "target-user-guid", Username: "target-user"},
			nil,
		)
	})

	When("neither origin nor client flag is provided", func() {
		BeforeEach(func() {
			cmd.Args.Organization = "some-org-name"
			cmd.Args.Role = flag.OrgRole{Role: "OrgAuditor"}
			cmd.Args.Username = "target-user-name"
		})

		It("creates the org role", func() {
			Expect(fakeActor.CreateOrgRoleCallCount()).To(Equal(1))
			givenRoleType, givenOrgGUID, givenUserName, givenOrigin, givenIsClient := fakeActor.CreateOrgRoleArgsForCall(0)
			Expect(givenRoleType).To(Equal(constant.OrgAuditorRole))
			Expect(givenOrgGUID).To(Equal("some-org-guid"))
			Expect(givenUserName).To(Equal("target-user-name"))
			Expect(givenOrigin).To(Equal(""))
			Expect(givenIsClient).To(BeFalse())
		})

		It("displays flavor text and returns without error", func() {
			Expect(testUI.Out).To(Say("Assigning role OrgAuditor to user target-user-name in org some-org-name as current-user..."))
			Expect(testUI.Out).To(Say("OK"))
			Expect(executeErr).NotTo(HaveOccurred())
		})
	})

	When("origin flag is provided", func() {
		BeforeEach(func() {
			cmd.Args.Organization = "some-org-name"
			cmd.Args.Role = flag.OrgRole{Role: "OrgAuditor"}
			cmd.Args.Username = "target-user-name"
			cmd.Origin = "ldap"
		})

		It("creates the org role by username", func() {
			Expect(fakeActor.CreateOrgRoleCallCount()).To(Equal(1))
			givenRoleType, givenOrgGUID, givenUserName, givenOrigin, givenIsClient := fakeActor.CreateOrgRoleArgsForCall(0)
			Expect(givenRoleType).To(Equal(constant.OrgAuditorRole))
			Expect(givenOrgGUID).To(Equal("some-org-guid"))
			Expect(givenUserName).To(Equal("target-user-name"))
			Expect(givenOrigin).To(Equal("ldap"))
			Expect(givenIsClient).To(BeFalse())
		})

		It("displays flavor text and returns without error", func() {
			Expect(testUI.Out).To(Say("Assigning role OrgAuditor to user target-user-name in org some-org-name as current-user..."))
			Expect(testUI.Out).To(Say("OK"))
			Expect(executeErr).NotTo(HaveOccurred())
		})
	})

	When("client flag is provided", func() {
		BeforeEach(func() {
			cmd.Args.Organization = "some-org-name"
			cmd.Args.Role = flag.OrgRole{Role: "OrgAuditor"}
			cmd.Args.Username = "target-user-name"
			cmd.IsClient = true
		})

		It("does not try to get the user", func() {
			Expect(fakeActor.GetUserCallCount()).To(Equal(0))
		})

		It("creates the org role correctly", func() {
			Expect(fakeActor.CreateOrgRoleCallCount()).To(Equal(1))
			givenRoleType, givenOrgGUID, givenUserGUID, givenOrigin, givenIsClient := fakeActor.CreateOrgRoleArgsForCall(0)
			Expect(givenRoleType).To(Equal(constant.OrgAuditorRole))
			Expect(givenOrgGUID).To(Equal("some-org-guid"))
			Expect(givenUserGUID).To(Equal("target-user-name"))
			Expect(givenOrigin).To(Equal(""))
			Expect(givenIsClient).To(BeTrue())
		})

		It("displays flavor text and returns without error", func() {
			Expect(testUI.Out).To(Say("Assigning role OrgAuditor to user target-user-name in org some-org-name as current-user..."))
			Expect(testUI.Out).To(Say("OK"))
			Expect(executeErr).NotTo(HaveOccurred())
		})
	})

	When("both client and origin flags are provided", func() {
		BeforeEach(func() {
			cmd.Args.Organization = "some-org-name"
			cmd.Args.Role = flag.OrgRole{Role: "OrgAuditor"}
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
			cmd.Args.Role = flag.OrgRole{Role: "OrgAuditor"}
			cmd.Args.Username = "target-user-name"

			fakeActor.CreateOrgRoleReturns(
				v7action.Warnings{"create-role-warning"},
				ccerror.RoleAlreadyExistsError{},
			)
		})

		It("displays warnings and returns without error", func() {
			Expect(testUI.Err).To(Say("create-role-warning"))
			Expect(testUI.Err).To(Say("User 'target-user-name' already has role 'OrgAuditor' in org 'some-org-name'."))
			Expect(testUI.Out).To(Say("OK"))
			Expect(executeErr).NotTo(HaveOccurred())
		})
	})

	When("when the role argument is invalid", func() {
		BeforeEach(func() {
			cmd.Args.Role = flag.OrgRole{Role: "MiddleManager"}
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError("Invalid role type."))
		})
	})

	When("getting the org fails", func() {
		BeforeEach(func() {
			cmd.Args.Role = flag.OrgRole{Role: "OrgAuditor"}

			fakeActor.GetOrganizationByNameReturns(
				resources.Organization{},
				v7action.Warnings{"get-org-warning"},
				errors.New("get-org-error"),
			)
		})

		It("displays warnings and returns an error", func() {
			Expect(testUI.Err).To(Say("get-org-warning"))
			Expect(executeErr).To(MatchError("get-org-error"))
		})
	})

	When("creating the role fails", func() {
		BeforeEach(func() {
			cmd.Args.Organization = "some-org-name"
			cmd.Args.Role = flag.OrgRole{Role: "OrgAuditor"}
			cmd.Args.Username = "target-user-name"

			fakeActor.CreateOrgRoleReturns(
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
