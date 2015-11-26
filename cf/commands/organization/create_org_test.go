package organization_test

import (
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"

	fakeflag "github.com/cloudfoundry/cli/cf/api/feature_flags/fakes"
	test_org "github.com/cloudfoundry/cli/cf/api/organizations/fakes"
	test_quota "github.com/cloudfoundry/cli/cf/api/quotas/fakes"
	userCmdFakes "github.com/cloudfoundry/cli/cf/commands/user/fakes"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	"github.com/cloudfoundry/cli/cf/command_registry"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("create-org command", func() {
	var (
		config              core_config.Repository
		ui                  *testterm.FakeUI
		requirementsFactory *testreq.FakeReqFactory
		orgRepo             *test_org.FakeOrganizationRepository
		quotaRepo           *test_quota.FakeQuotaRepository
		deps                command_registry.Dependency
		orgRoleSetter       *userCmdFakes.FakeOrgRoleSetter
		flagRepo            *fakeflag.FakeFeatureFlagRepository
		OriginalCommand     command_registry.Command
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.RepoLocator = deps.RepoLocator.SetOrganizationRepository(orgRepo)
		deps.RepoLocator = deps.RepoLocator.SetQuotaRepository(quotaRepo)
		deps.RepoLocator = deps.RepoLocator.SetFeatureFlagRepository(flagRepo)
		deps.Config = config

		//inject fake 'command dependency' into registry
		command_registry.Register(orgRoleSetter)

		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("create-org").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		config = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{}
		orgRepo = &test_org.FakeOrganizationRepository{}
		quotaRepo = &test_quota.FakeQuotaRepository{}
		flagRepo = &fakeflag.FakeFeatureFlagRepository{}
		config.SetApiVersion("2.36.9")

		orgRoleSetter = &userCmdFakes.FakeOrgRoleSetter{}
		//setup fakes to correctly interact with command_registry
		orgRoleSetter.SetDependencyStub = func(_ command_registry.Dependency, _ bool) command_registry.Command {
			return orgRoleSetter
		}
		orgRoleSetter.MetaDataReturns(command_registry.CommandMetadata{Name: "set-org-role"})

		//save original command and restore later
		OriginalCommand = command_registry.Commands.FindCommand("set-org-role")
	})

	AfterEach(func() {
		command_registry.Register(OriginalCommand)
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCliCommand("create-org", args, requirementsFactory, updateCommandDependency, false)
	}

	Describe("requirements", func() {
		It("fails with usage when not provided exactly one arg", func() {
			requirementsFactory.LoginSuccess = true
			runCommand()
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires an argument"},
			))
		})

		It("fails when not logged in", func() {
			Expect(runCommand("my-org")).To(BeFalse())
		})
	})

	Context("when logged in and provided the name of an org to create", func() {
		BeforeEach(func() {
			orgRepo.CreateReturns(nil)
			requirementsFactory.LoginSuccess = true
		})

		It("creates an org", func() {
			runCommand("my-org")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Creating org", "my-org", "my-user"},
				[]string{"OK"},
			))
			Expect(orgRepo.CreateArgsForCall(0).Name).To(Equal("my-org"))
		})

		It("fails and warns the user when the org already exists", func() {
			err := errors.NewHttpError(400, errors.ORG_EXISTS, "org already exists")
			orgRepo.CreateReturns(err)
			runCommand("my-org")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Creating org", "my-org"},
				[]string{"OK"},
				[]string{"my-org", "already exists"},
			))
		})

		Context("when CC api version supports assigning orgRole by name, and feature-flag 'set_roles_by_username' is enabled", func() {
			BeforeEach(func() {
				config.SetApiVersion("2.37.0")
				flagRepo.FindByNameReturns(models.FeatureFlag{
					Name:    "set_roles_by_username",
					Enabled: true,
				}, nil)
				orgRepo.FindByNameReturns(models.Organization{
					OrganizationFields: models.OrganizationFields{
						Guid: "my-org-guid",
					},
				}, nil)
			})

			It("assigns manager role to user", func() {
				runCommand("my-org")

				orgGuid, role, userGuid, userName := orgRoleSetter.SetOrgRoleArgsForCall(0)

				Ω(orgRoleSetter.SetOrgRoleCallCount()).To(Equal(1))
				Ω(orgGuid).To(Equal("my-org-guid"))
				Ω(role).To(Equal("OrgManager"))
				Ω(userGuid).To(Equal(""))
				Ω(userName).To(Equal("my-user"))
			})

			It("warns user about problem accessing feature-flag", func() {
				flagRepo.FindByNameReturns(models.FeatureFlag{}, errors.New("error error error"))

				runCommand("my-org")

				Ω(orgRoleSetter.SetOrgRoleCallCount()).To(Equal(0))
				Ω(ui.Outputs).To(ContainSubstrings(
					[]string{"Warning", "error error error"},
				))
			})

			It("fails on failing getting the guid of the newly created org", func() {
				orgRepo.FindByNameReturns(models.Organization{}, errors.New("cannot get org guid"))

				runCommand("my-org")

				Ω(orgRoleSetter.SetOrgRoleCallCount()).To(Equal(0))
				Ω(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"cannot get org guid"},
				))
			})

			It("fails on failing assigning org role to user", func() {
				orgRoleSetter.SetOrgRoleReturns(errors.New("failed to assign role"))

				runCommand("my-org")

				Ω(orgRoleSetter.SetOrgRoleCallCount()).To(Equal(1))
				Ω(ui.Outputs).To(ContainSubstrings(
					[]string{"Assigning role OrgManager to user my-user in org my-org ..."},
					[]string{"FAILED"},
					[]string{"failed to assign role"},
				))
			})
		})

		Context("when allowing a non-defualt quota", func() {
			var (
				quota models.QuotaFields
			)

			BeforeEach(func() {
				quota = models.QuotaFields{
					Name: "non-default-quota",
					Guid: "non-default-quota-guid",
				}
			})

			It("creates an org with a non-default quota", func() {
				quotaRepo.FindByNameReturns(quota, nil)
				runCommand("-q", "non-default-quota", "my-org")

				Expect(quotaRepo.FindByNameArgsForCall(0)).To(Equal("non-default-quota"))
				Expect(orgRepo.CreateArgsForCall(0).QuotaDefinition.Guid).To(Equal("non-default-quota-guid"))
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Creating org", "my-org"},
					[]string{"OK"},
				))
			})

			It("fails and warns the user when the quota cannot be found", func() {
				quotaRepo.FindByNameReturns(models.QuotaFields{}, errors.New("Could not find quota"))
				runCommand("-q", "non-default-quota", "my-org")

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Creating org", "my-org"},
					[]string{"Could not find quota"},
				))
			})
		})
	})
})
