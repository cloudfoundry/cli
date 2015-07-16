package organization_test

import (
	test_org "github.com/cloudfoundry/cli/cf/api/organizations/fakes"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("rename-org command", func() {
	var (
		requirementsFactory *testreq.FakeReqFactory
		orgRepo             *test_org.FakeOrganizationRepository
		ui                  *testterm.FakeUI
		configRepo          core_config.Repository
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.RepoLocator = deps.RepoLocator.SetOrganizationRepository(orgRepo)
		deps.Config = configRepo
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("rename-org").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		requirementsFactory = &testreq.FakeReqFactory{}
		orgRepo = &test_org.FakeOrganizationRepository{}
		ui = new(testterm.FakeUI)
		configRepo = testconfig.NewRepositoryWithDefaults()
	})

	var callRenameOrg = func(args []string) bool {
		return testcmd.RunCliCommand("rename-org", args, requirementsFactory, updateCommandDependency, false)
	}

	It("fails with usage when given less than two args", func() {
		callRenameOrg([]string{})
		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Incorrect Usage", "Requires", "arguments"},
		))

		callRenameOrg([]string{"foo"})
		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Incorrect Usage", "Requires", "arguments"},
		))
	})

	It("fails requirements when not logged in", func() {
		Expect(callRenameOrg([]string{"my-org", "my-new-org"})).To(BeFalse())
	})

	Context("when logged in and given an org to rename", func() {
		BeforeEach(func() {
			org := models.Organization{}
			org.Name = "the-old-org-name"
			org.Guid = "the-old-org-guid"
			requirementsFactory.Organization = org
			requirementsFactory.LoginSuccess = true
		})

		It("passes requirements", func() {
			Expect(callRenameOrg([]string{"the-old-org-name", "the-new-org-name"})).To(BeTrue())
		})

		It("renames an organization", func() {
			targetedOrgName := configRepo.OrganizationFields().Name
			callRenameOrg([]string{"the-old-org-name", "the-new-org-name"})
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Renaming org", "the-old-org-name", "the-new-org-name", "my-user"},
				[]string{"OK"},
			))

			guid, name := orgRepo.RenameArgsForCall(0)

			Expect(requirementsFactory.OrganizationName).To(Equal("the-old-org-name"))
			Expect(guid).To(Equal("the-old-org-guid"))
			Expect(name).To(Equal("the-new-org-name"))
			Expect(configRepo.OrganizationFields().Name).To(Equal(targetedOrgName))
		})

		Describe("when the organization is currently targeted", func() {
			It("updates the name of the org in the config", func() {
				configRepo.SetOrganizationFields(models.OrganizationFields{
					Guid: "the-old-org-guid",
					Name: "the-old-org-name",
				})
				callRenameOrg([]string{"the-old-org-name", "the-new-org-name"})
				Expect(configRepo.OrganizationFields().Name).To(Equal("the-new-org-name"))
			})
		})
	})
})
