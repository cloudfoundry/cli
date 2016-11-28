package organization_test

import (
	"os"

	"code.cloudfoundry.org/cli/cf/api/organizations/organizationsfakes"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	"code.cloudfoundry.org/cli/cf/trace/tracefakes"
	"code.cloudfoundry.org/cli/plugin/models"
	testcmd "code.cloudfoundry.org/cli/util/testhelpers/commands"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/cli/cf/commands/organization"
	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
)

var _ = Describe("orgs command", func() {
	var (
		ui                  *testterm.FakeUI
		orgRepo             *organizationsfakes.FakeOrganizationRepository
		configRepo          coreconfig.Repository
		requirementsFactory *requirementsfakes.FakeFactory
		deps                commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.Config = configRepo
		deps.RepoLocator = deps.RepoLocator.SetOrganizationRepository(orgRepo)
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("orgs").SetDependency(deps, pluginCall))
	}

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("orgs", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		orgRepo = new(organizationsfakes.FakeOrganizationRepository)
		requirementsFactory = new(requirementsfakes.FakeFactory)
		requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})

		deps = commandregistry.NewDependency(os.Stdout, new(tracefakes.FakePrinter), "")
	})

	Describe("requirements", func() {
		It("fails when not logged in", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Failing{Message: "not logged in"})

			Expect(runCommand()).To(BeFalse())
		})

		Context("when arguments are provided", func() {
			var cmd commandregistry.Command
			var flagContext flags.FlagContext

			BeforeEach(func() {
				cmd = &organization.ListOrgs{}
				cmd.SetDependency(deps, false)
				flagContext = flags.NewFlagContext(cmd.MetaData().Flags)
			})

			It("should fail with usage", func() {
				flagContext.Parse("blahblah")

				reqs, err := cmd.Requirements(requirementsFactory, flagContext)
				Expect(err).NotTo(HaveOccurred())

				err = testcmd.RunRequirements(reqs)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Incorrect Usage"))
				Expect(err.Error()).To(ContainSubstring("No argument required"))
			})
		})
	})

	Describe("when invoked by a plugin", func() {
		var (
			pluginOrgsModel []plugin_models.GetOrgs_Model
		)

		BeforeEach(func() {
			org1 := models.Organization{}
			org1.Name = "Organization-1"
			org1.GUID = "org-1-guid"

			org2 := models.Organization{}
			org2.Name = "Organization-2"

			org3 := models.Organization{}
			org3.Name = "Organization-3"

			orgRepo.ListOrgsReturns([]models.Organization{org1, org2, org3}, nil)

			pluginOrgsModel = []plugin_models.GetOrgs_Model{}
			deps.PluginModels.Organizations = &pluginOrgsModel
		})

		It("populates the plugin models upon execution", func() {
			testcmd.RunCLICommand("orgs", []string{}, requirementsFactory, updateCommandDependency, true, ui)
			Expect(pluginOrgsModel[0].Name).To(Equal("Organization-1"))
			Expect(pluginOrgsModel[0].Guid).To(Equal("org-1-guid"))
			Expect(pluginOrgsModel[1].Name).To(Equal("Organization-2"))
			Expect(pluginOrgsModel[2].Name).To(Equal("Organization-3"))
		})
	})

	Context("when there are orgs to be listed", func() {
		BeforeEach(func() {
			org1 := models.Organization{}
			org1.Name = "Organization-1"

			org2 := models.Organization{}
			org2.Name = "Organization-2"

			org3 := models.Organization{}
			org3.Name = "Organization-3"

			orgRepo.ListOrgsReturns([]models.Organization{org1, org2, org3}, nil)
		})

		It("tries to get the organizations", func() {
			runCommand()
			Expect(orgRepo.ListOrgsCallCount()).To(Equal(1))
			Expect(orgRepo.ListOrgsArgsForCall(0)).To(Equal(0))
		})

		It("lists orgs", func() {
			runCommand()

			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Getting orgs as my-user"},
				[]string{"Organization-1"},
				[]string{"Organization-2"},
				[]string{"Organization-3"},
			))
		})
	})

	It("tells the user when no orgs were found", func() {
		orgRepo.ListOrgsReturns([]models.Organization{}, nil)
		runCommand()

		Expect(ui.Outputs()).To(ContainSubstrings(
			[]string{"Getting orgs as my-user"},
			[]string{"No orgs found"},
		))
	})
})
