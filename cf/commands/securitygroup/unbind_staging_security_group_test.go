package securitygroup_test

import (
	"code.cloudfoundry.org/cli/cf/api/securitygroups/defaults/staging/stagingfakes"
	"code.cloudfoundry.org/cli/cf/api/securitygroups/securitygroupsfakes"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	testcmd "code.cloudfoundry.org/cli/util/testhelpers/commands"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"

	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("unbind-staging-security-group command", func() {
	var (
		ui                            *testterm.FakeUI
		configRepo                    coreconfig.Repository
		requirementsFactory           *requirementsfakes.FakeFactory
		fakeSecurityGroupRepo         *securitygroupsfakes.FakeSecurityGroupRepo
		fakeStagingSecurityGroupsRepo *stagingfakes.FakeSecurityGroupsRepo
		deps                          commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.RepoLocator = deps.RepoLocator.SetSecurityGroupRepository(fakeSecurityGroupRepo)
		deps.RepoLocator = deps.RepoLocator.SetStagingSecurityGroupRepository(fakeStagingSecurityGroupsRepo)
		deps.Config = configRepo
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("unbind-staging-security-group").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = new(requirementsfakes.FakeFactory)
		fakeSecurityGroupRepo = new(securitygroupsfakes.FakeSecurityGroupRepo)
		fakeStagingSecurityGroupsRepo = new(stagingfakes.FakeSecurityGroupsRepo)
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("unbind-staging-security-group", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	Describe("requirements", func() {
		It("fails when the user is not logged in", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Failing{Message: "not logged in"})
			Expect(runCommand("name")).To(BeFalse())
		})

		It("fails with usage when a name is not provided", func() {
			runCommand()
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "argument"},
			))
		})
	})

	Context("when the user is logged in and provides the name of a group", func() {
		BeforeEach(func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
		})

		Context("security group exists", func() {
			BeforeEach(func() {
				group := models.SecurityGroup{}
				group.GUID = "just-pretend-this-is-a-guid"
				group.Name = "a-security-group-name"
				fakeSecurityGroupRepo.ReadReturns(group, nil)
			})

			JustBeforeEach(func() {
				runCommand("a-security-group-name")
			})

			It("unbinds the group from the default staging group set", func() {
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Unbinding", "security group", "a-security-group-name", "my-user"},
					[]string{"OK"},
				))

				Expect(fakeSecurityGroupRepo.ReadArgsForCall(0)).To(Equal("a-security-group-name"))
				Expect(fakeStagingSecurityGroupsRepo.UnbindFromStagingSetArgsForCall(0)).To(Equal("just-pretend-this-is-a-guid"))
			})
		})

		Context("when the security group does not exist", func() {
			BeforeEach(func() {
				fakeSecurityGroupRepo.ReadReturns(models.SecurityGroup{}, errors.NewModelNotFoundError("security group", "anana-qui-parle"))
			})

			It("warns the user", func() {
				runCommand("anana-qui-parle")
				Expect(ui.WarnOutputs).To(ContainSubstrings(
					[]string{"Security group", "anana-qui-parle", "does not exist"},
				))

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"OK"},
				))
			})
		})
	})
})
