package organization_test

import (
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/plugin/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func callShowOrg(args []string, requirementsFactory *testreq.FakeReqFactory) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)

	token := core_config.TokenInfo{Username: "my-user"}

	spaceFields := models.SpaceFields{}
	spaceFields.Name = "my-space"

	orgFields := models.OrganizationFields{}
	orgFields.Name = "my-org"

	configRepo := testconfig.NewRepositoryWithAccessToken(token)
	configRepo.SetSpaceFields(spaceFields)
	configRepo.SetOrganizationFields(orgFields)

	return
}

var _ = Describe("org command", func() {
	var (
		ui                  *testterm.FakeUI
		configRepo          core_config.Repository
		requirementsFactory *testreq.FakeReqFactory
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.Config = configRepo
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("org").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		requirementsFactory = &testreq.FakeReqFactory{}
		configRepo = testconfig.NewRepositoryWithDefaults()

		deps = command_registry.NewDependency()
		updateCommandDependency(false)
	})

	runCommand := func(args ...string) bool {
		cmd := command_registry.Commands.FindCommand("org")
		return testcmd.RunCliCommand(cmd, args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("fails when not logged in", func() {
			Expect(runCommand("whoops")).To(BeFalse())
		})

		It("fails with usage when not provided exactly one arg", func() {
			requirementsFactory.LoginSuccess = true
			runCommand("too", "much")
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires an argument"},
			))
		})
	})

	Context("when logged in, and provided the name of an org", func() {
		BeforeEach(func() {
			developmentSpaceFields := models.SpaceFields{}
			developmentSpaceFields.Name = "development"
			stagingSpaceFields := models.SpaceFields{}
			stagingSpaceFields.Name = "staging"
			domainFields := models.DomainFields{}
			domainFields.Name = "cfapps.io"
			cfAppDomainFields := models.DomainFields{}
			cfAppDomainFields.Name = "cf-app.com"

			org := models.Organization{}
			org.Name = "my-org"
			org.Guid = "my-org-guid"
			org.QuotaDefinition = models.NewQuotaFields("cantina-quota", 512, 256, 2, 5, true)
			org.Spaces = []models.SpaceFields{developmentSpaceFields, stagingSpaceFields}
			org.Domains = []models.DomainFields{domainFields, cfAppDomainFields}
			org.SpaceQuotas = []models.SpaceQuota{
				{Name: "space-quota-1"},
				{Name: "space-quota-2"},
			}

			requirementsFactory.LoginSuccess = true
			requirementsFactory.Organization = org
		})

		It("shows the org with the given name", func() {
			runCommand("my-org")

			Expect(requirementsFactory.OrganizationName).To(Equal("my-org"))
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Getting info for org", "my-org", "my-user"},
				[]string{"OK"},
				[]string{"my-org"},
				[]string{"domains:", "cfapps.io", "cf-app.com"},
				[]string{"quota: ", "cantina-quota", "512M", "256M instance memory limit", "2 routes", "5 services", "paid services allowed"},
				[]string{"spaces:", "development", "staging"},
				[]string{"space quotas:", "space-quota-1", "space-quota-2"},
			))
		})

		Context("when the guid flag is provided", func() {
			It("shows only the org guid", func() {
				runCommand("--guid", "my-org")

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"my-org-guid"},
				))

				Expect(ui.Outputs).ToNot(ContainSubstrings(
					[]string{"Getting info for org", "my-org", "my-user"},
				))
			})
		})

		Context("when invoked by a plugin", func() {
			var (
				pluginModel plugin_models.Organization
			)
			BeforeEach(func() {
				pluginModel = plugin_models.Organization{}
				deps.PluginModels.Organization = &pluginModel
				updateCommandDependency(true)
			})

			It("populates the plugin model", func() {
				runCommand("my-org")

				Ω(pluginModel.Name).To(Equal("my-org"))
				Ω(pluginModel.Guid).To(Equal("my-org-guid"))
				Ω(pluginModel.QuotaDefinition.Name).To(Equal("cantina-quota"))
				Ω(pluginModel.QuotaDefinition.MemoryLimit).To(Equal(int64(512)))
				Ω(pluginModel.QuotaDefinition.InstanceMemoryLimit).To(Equal(int64(256)))
				Ω(pluginModel.QuotaDefinition.RoutesLimit).To(Equal(2))
				Ω(pluginModel.QuotaDefinition.ServicesLimit).To(Equal(5))
				Ω(pluginModel.QuotaDefinition.NonBasicServicesAllowed).To(BeTrue())
			})
		})
	})
})
