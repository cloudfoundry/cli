package v7_test

import (
	"code.cloudfoundry.org/cli/actor/v7action"
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

var _ = Describe("unset-org-role Command", func() {
	var (
		cmd             UnsetOrgRoleCommand
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

		cmd = UnsetOrgRoleCommand{
			BaseCommand: command.BaseCommand{
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

		fakeActor.GetOrganizationByNameReturns(
			v7action.Organization{GUID: "some-org-guid", Name: "some-org-name"},
			v7action.Warnings{"get-org-warning"},
			nil,
		)
	})

	When("neither origin nor client flag is provided", func() {
		BeforeEach(func() {
			cmd.Args.Organization = "some-org-name"
			cmd.Args.Role = flag.OrgRole{Role: "OrgAuditor"}
			cmd.Args.Username = "target-user-name"
		})

		It("deletes the org role", func() {
			Expect(fakeActor.DeleteOrgRoleCallCount()).To(Equal(1))
			givenRoleType, givenOrgGUID, givenUserName, givenOrigin, givenIsClient := fakeActor.DeleteOrgRoleArgsForCall(0)
			Expect(givenRoleType).To(Equal(constant.OrgAuditorRole))
			Expect(givenOrgGUID).To(Equal("some-org-guid"))
			Expect(givenUserName).To(Equal("target-user-name"))
			Expect(givenOrigin).To(Equal("uaa"))
			Expect(givenIsClient).To(BeFalse())
		})

		It("displays flavor text and returns without error", func() {
			Expect(testUI.Out).To(Say("Removing role OrgAuditor from user target-user-name in org some-org-name as current-user..."))
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

		It("deletes the org role", func() {
			Expect(fakeActor.DeleteOrgRoleCallCount()).To(Equal(1))
			givenRoleType, givenOrgGUID, givenUserName, givenOrigin, givenIsClient := fakeActor.DeleteOrgRoleArgsForCall(0)
			Expect(givenRoleType).To(Equal(constant.OrgAuditorRole))
			Expect(givenOrgGUID).To(Equal("some-org-guid"))
			Expect(givenUserName).To(Equal("target-user-name"))
			Expect(givenOrigin).To(Equal("ldap"))
			Expect(givenIsClient).To(BeFalse())
		})

		It("displays flavor text and returns without error", func() {
			Expect(testUI.Out).To(Say("Removing role OrgAuditor from user target-user-name in org some-org-name as current-user..."))
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

		It("deletes the org role correctly", func() {
			givenRoleType, givenOrgGUID, givenUserName, givenOrigin, givenIsClient := fakeActor.DeleteOrgRoleArgsForCall(0)
			Expect(givenRoleType).To(Equal(constant.OrgAuditorRole))
			Expect(givenOrgGUID).To(Equal("some-org-guid"))
			Expect(givenUserName).To(Equal("target-user-name"))
			Expect(givenOrigin).To(Equal("uaa"))
			Expect(givenIsClient).To(BeTrue())
		})

		It("displays flavor text and returns without error", func() {
			Expect(testUI.Out).To(Say("Removing role OrgAuditor from user target-user-name in org some-org-name as current-user..."))
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

	When("the role does not exist", func() {
		BeforeEach(func() {
			cmd.Args.Organization = "some-org-name"
			cmd.Args.Role = flag.OrgRole{Role: "OrgAuditor"}
			cmd.Args.Username = "target-user-name"

			fakeActor.DeleteOrgRoleReturns(
				v7action.Warnings{"delete-role-warning"},
				nil,
			)
		})

		It("displays warnings and returns without error", func() {
			Expect(testUI.Err).To(Say("delete-role-warning"))
			Expect(testUI.Out).To(Say("OK"))
			Expect(executeErr).NotTo(HaveOccurred())
		})
	})

	When("invalid role arg is given", func() {
		BeforeEach(func() {
			cmd.Args.Role = flag.OrgRole{Role: "Astronaut"}
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError("Invalid role type."))
		})
	})

	When("getting the org fails", func() {
		BeforeEach(func() {
			cmd.Args.Role = flag.OrgRole{Role: "OrgAuditor"}

			fakeActor.GetOrganizationByNameReturns(
				v7action.Organization{},
				v7action.Warnings{"get-org-warning"},
				errors.New("get-org-error"),
			)
		})

		It("displays warnings and returns an error", func() {
			Expect(testUI.Err).To(Say("get-org-warning"))
			Expect(executeErr).To(MatchError("get-org-error"))
		})
	})

	When("deleting the org role fails", func() {
		BeforeEach(func() {
			cmd.Args.Organization = "some-org-name"
			cmd.Args.Role = flag.OrgRole{Role: "OrgAuditor"}
			cmd.Args.Username = "target-user-name"

			fakeActor.DeleteOrgRoleReturns(
				v7action.Warnings{"delete-role-warning"},
				errors.New("delete-role-error"),
			)
		})

		It("displays warnings and returns an error", func() {
			Expect(testUI.Err).To(Say("delete-role-warning"))
			Expect(executeErr).To(MatchError("delete-role-error"))
		})
	})
})
