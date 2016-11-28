package organization_test

import (
	"code.cloudfoundry.org/cli/cf/api/organizations/organizationsfakes"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	testcmd "code.cloudfoundry.org/cli/util/testhelpers/commands"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
)

var _ = Describe("rename-org command", func() {
	var (
		requirementsFactory *requirementsfakes.FakeFactory
		orgRepo             *organizationsfakes.FakeOrganizationRepository
		ui                  *testterm.FakeUI
		configRepo          coreconfig.Repository
		deps                commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.RepoLocator = deps.RepoLocator.SetOrganizationRepository(orgRepo)
		deps.Config = configRepo
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("rename-org").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		requirementsFactory = new(requirementsfakes.FakeFactory)
		orgRepo = new(organizationsfakes.FakeOrganizationRepository)
		ui = new(testterm.FakeUI)
		configRepo = testconfig.NewRepositoryWithDefaults()
	})

	var callRenameOrg = func(args []string) bool {
		return testcmd.RunCLICommand("rename-org", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	It("fails with usage when given less than two args", func() {
		callRenameOrg([]string{})
		Expect(ui.Outputs()).To(ContainSubstrings(
			[]string{"Incorrect Usage", "Requires", "arguments"},
		))

		callRenameOrg([]string{"foo"})
		Expect(ui.Outputs()).To(ContainSubstrings(
			[]string{"Incorrect Usage", "Requires", "arguments"},
		))
	})

	It("fails requirements when not logged in", func() {
		requirementsFactory.NewLoginRequirementReturns(requirements.Failing{Message: "not logged in"})
		Expect(callRenameOrg([]string{"my-org", "my-new-org"})).To(BeFalse())
	})

	Context("when logged in and given an org to rename", func() {
		BeforeEach(func() {
			org := models.Organization{}
			org.Name = "the-old-org-name"
			org.GUID = "the-old-org-guid"
			orgReq := new(requirementsfakes.FakeOrganizationRequirement)
			orgReq.GetOrganizationReturns(org)
			requirementsFactory.NewOrganizationRequirementReturns(orgReq)
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
		})

		It("passes requirements", func() {
			Expect(callRenameOrg([]string{"the-old-org-name", "the-new-org-name"})).To(BeTrue())
		})

		It("renames an organization", func() {
			targetedOrgName := configRepo.OrganizationFields().Name
			callRenameOrg([]string{"the-old-org-name", "the-new-org-name"})
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Renaming org", "the-old-org-name", "the-new-org-name", "my-user"},
				[]string{"OK"},
			))

			guid, name := orgRepo.RenameArgsForCall(0)

			Expect(guid).To(Equal("the-old-org-guid"))
			Expect(name).To(Equal("the-new-org-name"))
			Expect(configRepo.OrganizationFields().Name).To(Equal(targetedOrgName))
		})

		Describe("when the organization is currently targeted", func() {
			It("updates the name of the org in the config", func() {
				configRepo.SetOrganizationFields(models.OrganizationFields{
					GUID: "the-old-org-guid",
					Name: "the-old-org-name",
				})
				callRenameOrg([]string{"the-old-org-name", "the-new-org-name"})
				Expect(configRepo.OrganizationFields().Name).To(Equal("the-new-org-name"))
			})
		})
	})
})
