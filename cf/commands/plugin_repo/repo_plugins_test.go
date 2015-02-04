package plugin_repo_test

import (
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"

	. "github.com/cloudfoundry/cli/cf/commands/plugin_repo"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"

	fakes "github.com/cloudfoundry/cli/cf/actors/plugin_repo/fakes"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	clipr "github.com/cloudfoundry-incubator/cli-plugin-repo/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("repo-plugins", func() {
	var (
		ui                  *testterm.FakeUI
		config              core_config.ReadWriter
		requirementsFactory *testreq.FakeReqFactory
		fakePluginRepo      *fakes.FakePluginRepo
	)

	BeforeEach(func() {
		fakePluginRepo = &fakes.FakePluginRepo{}
		ui = &testterm.FakeUI{}
		requirementsFactory = &testreq.FakeReqFactory{}
		config = testconfig.NewRepositoryWithDefaults()

		config.SetPluginRepo(models.PluginRepo{
			Name: "repo1",
			Url:  "",
		})

		config.SetPluginRepo(models.PluginRepo{
			Name: "repo2",
			Url:  "",
		})
	})

	var callRepoPlugins = func(args ...string) bool {
		cmd := NewRepoPlugins(ui, config, fakePluginRepo)
		return testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	Context("If repo name is provided by '-r'", func() {
		It("list plugins from just the named repo", func() {
			callRepoPlugins("-r", "repo2")

			Ω(fakePluginRepo.GetPluginsArgsForCall(0)[0].Name).To(Equal("repo2"))
			Ω(len(fakePluginRepo.GetPluginsArgsForCall(0))).To(Equal(1))

			Ω(ui.Outputs).To(ContainSubstrings([]string{"Getting plugins from repository 'repo2'"}))
		})
	})

	Context("If no repo name is provided", func() {
		It("list plugins from just the named repo", func() {
			callRepoPlugins()

			Ω(fakePluginRepo.GetPluginsArgsForCall(0)[0].Name).To(Equal("repo1"))
			Ω(len(fakePluginRepo.GetPluginsArgsForCall(0))).To(Equal(2))
			Ω(fakePluginRepo.GetPluginsArgsForCall(0)[1].Name).To(Equal("repo2"))

			Ω(ui.Outputs).To(ContainSubstrings([]string{"Getting plugins from all repositories"}))
		})
	})

	Context("when GetPlugins returns a list of plugin meta data", func() {
		It("lists all plugin data", func() {
			result := make(map[string][]clipr.Plugin)
			result["repo1"] = []clipr.Plugin{
				clipr.Plugin{
					Name:        "plugin1",
					Description: "none1",
				},
			}
			result["repo2"] = []clipr.Plugin{
				clipr.Plugin{
					Name:        "plugin2",
					Description: "none2",
				},
			}
			fakePluginRepo.GetPluginsReturns(result, []string{})

			callRepoPlugins()

			Ω(ui.Outputs).ToNot(ContainSubstrings([]string{"Logged errors:"}))
			Ω(ui.Outputs).To(ContainSubstrings([]string{"repo1"}))
			Ω(ui.Outputs).To(ContainSubstrings([]string{"plugin1"}))
			Ω(ui.Outputs).To(ContainSubstrings([]string{"repo2"}))
			Ω(ui.Outputs).To(ContainSubstrings([]string{"plugin2"}))
		})
	})

	Context("If errors are reported back from GetPlugins()", func() {
		It("informs user about the errors", func() {
			fakePluginRepo.GetPluginsReturns(nil, []string{
				"error from repo1",
				"error from repo2",
			})

			callRepoPlugins()

			Ω(ui.Outputs).To(ContainSubstrings([]string{"Logged errors:"}))
			Ω(ui.Outputs).To(ContainSubstrings([]string{"error from repo1"}))
			Ω(ui.Outputs).To(ContainSubstrings([]string{"error from repo2"}))
		})
	})

})
