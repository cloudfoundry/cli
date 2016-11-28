package pluginrepo_test

import (
	testcmd "code.cloudfoundry.org/cli/util/testhelpers/commands"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"

	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("delte-plugin-repo", func() {
	var (
		ui                  *testterm.FakeUI
		config              coreconfig.Repository
		requirementsFactory *requirementsfakes.FakeFactory
		deps                commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.Config = config
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("remove-plugin-repo").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		requirementsFactory = new(requirementsfakes.FakeFactory)
		config = testconfig.NewRepositoryWithDefaults()
	})

	var callRemovePluginRepo = func(args ...string) bool {
		return testcmd.RunCLICommand("remove-plugin-repo", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	Context("When repo name is valid", func() {
		BeforeEach(func() {
			config.SetPluginRepo(models.PluginRepo{
				Name: "repo1",
				URL:  "http://someserver1.com:1234",
			})

			config.SetPluginRepo(models.PluginRepo{
				Name: "repo2",
				URL:  "http://server2.org:8080",
			})
		})

		It("deletes the repo from the config", func() {
			callRemovePluginRepo("repo1")
			Expect(len(config.PluginRepos())).To(Equal(1))
			Expect(config.PluginRepos()[0].Name).To(Equal("repo2"))
			Expect(config.PluginRepos()[0].URL).To(Equal("http://server2.org:8080"))
		})
	})

	Context("When named repo doesn't exist", func() {
		BeforeEach(func() {
			config.SetPluginRepo(models.PluginRepo{
				Name: "repo1",
				URL:  "http://someserver1.com:1234",
			})

			config.SetPluginRepo(models.PluginRepo{
				Name: "repo2",
				URL:  "http://server2.org:8080",
			})
		})

		It("doesn't change the config the config", func() {
			callRemovePluginRepo("fake-repo")

			Expect(len(config.PluginRepos())).To(Equal(2))
			Expect(config.PluginRepos()[0].Name).To(Equal("repo1"))
			Expect(config.PluginRepos()[0].URL).To(Equal("http://someserver1.com:1234"))
			Expect(ui.Outputs()).To(ContainSubstrings([]string{"fake-repo", "does not exist as a repo"}))
		})
	})
})
