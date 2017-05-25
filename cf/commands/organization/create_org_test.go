package organization_test

import (
	"code.cloudfoundry.org/cli/cf/commands/user/userfakes"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"

	"code.cloudfoundry.org/cli/cf/api/featureflags/featureflagsfakes"
	"code.cloudfoundry.org/cli/cf/api/organizations/organizationsfakes"
	"code.cloudfoundry.org/cli/cf/api/quotas/quotasfakes"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	testcmd "code.cloudfoundry.org/cli/util/testhelpers/commands"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"

	"code.cloudfoundry.org/cli/cf/commandregistry"
	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("create-org command", func() {
	var (
		config              coreconfig.Repository
		ui                  *testterm.FakeUI
		requirementsFactory *requirementsfakes.FakeFactory
		orgRepo             *organizationsfakes.FakeOrganizationRepository
		quotaRepo           *quotasfakes.FakeQuotaRepository
		deps                commandregistry.Dependency
		orgRoleSetter       *userfakes.FakeOrgRoleSetter
		flagRepo            *featureflagsfakes.FakeFeatureFlagRepository
		OriginalCommand     commandregistry.Command
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.RepoLocator = deps.RepoLocator.SetOrganizationRepository(orgRepo)
		deps.RepoLocator = deps.RepoLocator.SetQuotaRepository(quotaRepo)
		deps.RepoLocator = deps.RepoLocator.SetFeatureFlagRepository(flagRepo)
		deps.Config = config

		//inject fake 'command dependency' into registry
		commandregistry.Register(orgRoleSetter)

		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("create-org").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		config = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = new(requirementsfakes.FakeFactory)
		requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
		orgRepo = new(organizationsfakes.FakeOrganizationRepository)
		quotaRepo = new(quotasfakes.FakeQuotaRepository)
		flagRepo = new(featureflagsfakes.FakeFeatureFlagRepository)
		config.SetAPIVersion("2.36.9")

		orgRoleSetter = new(userfakes.FakeOrgRoleSetter)
		//setup fakes to correctly interact with commandregistry
		orgRoleSetter.SetDependencyStub = func(_ commandregistry.Dependency, _ bool) commandregistry.Command {
			return orgRoleSetter
		}
		orgRoleSetter.MetaDataReturns(commandregistry.CommandMetadata{Name: "set-org-role"})

		//save original command and restore later
		OriginalCommand = commandregistry.Commands.FindCommand("set-org-role")
	})

	AfterEach(func() {
		commandregistry.Register(OriginalCommand)
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("create-org", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	Describe("requirements", func() {
		It("fails with usage when not provided exactly one arg", func() {
			runCommand()
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires an argument"},
			))
		})

		It("fails when not logged in", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Failing{Message: "not logged in"})
			Expect(runCommand("my-org")).To(BeFalse())
		})
	})

	Context("when logged in and provided the name of an org to create", func() {
		BeforeEach(func() {
			orgRepo.CreateReturns(nil)
		})

		It("creates an org", func() {
			runCommand("my-org")

			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Creating org", "my-org", "my-user"},
				[]string{"OK"},
				[]string{`TIP: Use 'cf target -o "my-org"' to target new org`},
			))
			Expect(orgRepo.CreateArgsForCall(0).Name).To(Equal("my-org"))
		})

		It("fails and warns the user when the org already exists", func() {
			err := errors.NewHTTPError(400, errors.OrganizationNameTaken, "org already exists")
			orgRepo.CreateReturns(err)
			runCommand("my-org")

			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Creating org", "my-org"},
				[]string{"OK"},
				[]string{"my-org", "already exists"},
			))

			Expect(ui.Outputs()).NotTo(ContainSubstrings(
				[]string{`TIP: Use 'cf target -o "my-org"' to target new org`},
			))
		})

		Context("when CC api version supports assigning orgRole by name, and feature-flag 'set_roles_by_username' is enabled", func() {
			BeforeEach(func() {
				config.SetAPIVersion("2.37.0")
				flagRepo.FindByNameReturns(models.FeatureFlag{
					Name:    "set_roles_by_username",
					Enabled: true,
				}, nil)
				orgRepo.FindByNameReturns(models.Organization{
					OrganizationFields: models.OrganizationFields{
						GUID: "my-org-guid",
					},
				}, nil)
			})

			It("assigns manager role to user", func() {
				runCommand("my-org")

				orgGUID, role, userGUID, userName := orgRoleSetter.SetOrgRoleArgsForCall(0)

				Expect(orgRoleSetter.SetOrgRoleCallCount()).To(Equal(1))
				Expect(orgGUID).To(Equal("my-org-guid"))
				Expect(role).To(Equal(models.RoleOrgManager))
				Expect(userGUID).To(Equal(""))
				Expect(userName).To(Equal("my-user"))
			})

			It("warns user about problem accessing feature-flag", func() {
				flagRepo.FindByNameReturns(models.FeatureFlag{}, errors.New("error error error"))

				runCommand("my-org")

				Expect(orgRoleSetter.SetOrgRoleCallCount()).To(Equal(0))
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Warning", "error error error"},
				))

			})

			It("fails on failing getting the guid of the newly created org", func() {
				orgRepo.FindByNameReturns(models.Organization{}, errors.New("cannot get org guid"))

				runCommand("my-org")

				Expect(orgRoleSetter.SetOrgRoleCallCount()).To(Equal(0))
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"cannot get org guid"},
				))

			})

			It("fails on failing assigning org role to user", func() {
				orgRoleSetter.SetOrgRoleReturns(errors.New("failed to assign role"))

				runCommand("my-org")

				Expect(orgRoleSetter.SetOrgRoleCallCount()).To(Equal(1))
				Expect(ui.Outputs()).To(ContainSubstrings(
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
					GUID: "non-default-quota-guid",
				}
			})

			It("creates an org with a non-default quota", func() {
				quotaRepo.FindByNameReturns(quota, nil)
				runCommand("-q", "non-default-quota", "my-org")

				Expect(quotaRepo.FindByNameArgsForCall(0)).To(Equal("non-default-quota"))
				Expect(orgRepo.CreateArgsForCall(0).QuotaDefinition.GUID).To(Equal("non-default-quota-guid"))
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Creating org", "my-org"},
					[]string{"OK"},
				))
			})

			It("fails and warns the user when the quota cannot be found", func() {
				quotaRepo.FindByNameReturns(models.QuotaFields{}, errors.New("Could not find quota"))
				runCommand("-q", "non-default-quota", "my-org")

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Creating org", "my-org"},
					[]string{"Could not find quota"},
				))
			})
		})
	})
})
