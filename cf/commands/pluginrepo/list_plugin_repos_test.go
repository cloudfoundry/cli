package pluginrepo_test

import (
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"

	testcmd "code.cloudfoundry.org/cli/util/testhelpers/commands"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"

	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"

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
