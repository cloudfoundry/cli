package pluginrepo_test

import (
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements/requirementsfakes"

	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("list-plugin-repo", func() {
	var (
		ui                  *testterm.FakeUI
		config              coreconfig.Repository
		requirementsFactory *requirementsfakes.FakeFactory
		deps                commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.Config = config
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("list-plugin-repos").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		requirementsFactory = new(requirementsfakes.FakeFactory)
		config = testconfig.NewRepositoryWithDefaults()
	})

	var callListPluginRepos = func(args ...string) bool {
		return testcmd.RunCLICommand("list-plugin-repos", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	It("lists all added plugin repo in a table", func() {
		config.SetPluginRepo(models.PluginRepo{
			Name: "repo1",
			URL:  "http://url1.com",
		})
		config.SetPluginRepo(models.PluginRepo{
			Name: "repo2",
			URL:  "http://url2.com",
		})

		callListPluginRepos()

		Expect(ui.Outputs()).To(ContainSubstrings(
			[]string{"repo1", "http://url1.com"},
			[]string{"repo2", "http://url2.com"},
		))

	})

})
