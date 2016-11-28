package pluginrepo_test

import (
	"code.cloudfoundry.org/cli/cf/actors/pluginrepo/pluginrepofakes"
	"code.cloudfoundry.org/cli/cf/commands/pluginrepo"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"

	"code.cloudfoundry.org/cli/cf/commandregistry"
	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"

	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"

	clipr "github.com/cloudfoundry-incubator/cli-plugin-repo/web"

	"code.cloudfoundry.org/cli/cf/flags"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("repo-plugins", func() {
	var (
		ui                  *testterm.FakeUI
		config              coreconfig.Repository
		requirementsFactory *requirementsfakes.FakeFactory
		fakePluginRepo      *pluginrepofakes.FakePluginRepo
		deps                commandregistry.Dependency
		cmd                 *pluginrepo.RepoPlugins
		flagContext         flags.FlagContext
	)

	BeforeEach(func() {
		fakePluginRepo = new(pluginrepofakes.FakePluginRepo)
		ui = &testterm.FakeUI{}
		requirementsFactory = new(requirementsfakes.FakeFactory)
		config = testconfig.NewRepositoryWithDefaults()

		deps = commandregistry.Dependency{
			UI:         ui,
			Config:     config,
			PluginRepo: fakePluginRepo,
		}

		cmd = new(pluginrepo.RepoPlugins)
		cmd = cmd.SetDependency(deps, false).(*pluginrepo.RepoPlugins)
		flagContext = flags.NewFlagContext(cmd.MetaData().Flags)
	})

	var callRepoPlugins = func(args ...string) error {
		err := flagContext.Parse(args...)
		if err != nil {
			return err
		}

		cmd.Execute(flagContext)
		return nil
	}

	Context("when using the default CF-Community repo", func() {
		BeforeEach(func() {
			config.SetPluginRepo(models.PluginRepo{
				Name: "cf",
				URL:  "http://plugins.cloudfoundry.org",
			})
		})

		It("uses https when pointing to plugins.cloudfoundry.org", func() {
			err := callRepoPlugins()
			Expect(err).NotTo(HaveOccurred())

			Expect(fakePluginRepo.GetPluginsCallCount()).To(Equal(1))
			Expect(fakePluginRepo.GetPluginsArgsForCall(0)[0].Name).To(Equal("cf"))
			Expect(fakePluginRepo.GetPluginsArgsForCall(0)[0].URL).To(Equal("https://plugins.cloudfoundry.org"))
			Expect(len(fakePluginRepo.GetPluginsArgsForCall(0))).To(Equal(1))
		})
	})

	Context("when using other plugin repos", func() {
		BeforeEach(func() {
			config.SetPluginRepo(models.PluginRepo{
				Name: "repo1",
				URL:  "",
			})

			config.SetPluginRepo(models.PluginRepo{
				Name: "repo2",
				URL:  "",
			})
		})

		Context("If repo name is provided by '-r'", func() {
			It("list plugins from just the named repo", func() {
				err := callRepoPlugins("-r", "repo2")
				Expect(err).NotTo(HaveOccurred())

				Expect(fakePluginRepo.GetPluginsArgsForCall(0)[0].Name).To(Equal("repo2"))
				Expect(len(fakePluginRepo.GetPluginsArgsForCall(0))).To(Equal(1))

				Expect(ui.Outputs()).To(ContainSubstrings([]string{"Getting plugins from repository 'repo2'"}))
			})
		})

		Context("If no repo name is provided", func() {
			It("list plugins from just the named repo", func() {
				err := callRepoPlugins()
				Expect(err).NotTo(HaveOccurred())

				Expect(fakePluginRepo.GetPluginsArgsForCall(0)[0].Name).To(Equal("repo1"))
				Expect(len(fakePluginRepo.GetPluginsArgsForCall(0))).To(Equal(2))
				Expect(fakePluginRepo.GetPluginsArgsForCall(0)[1].Name).To(Equal("repo2"))

				Expect(ui.Outputs()).To(ContainSubstrings([]string{"Getting plugins from all repositories"}))
			})
		})

		Context("when GetPlugins returns a list of plugin meta data", func() {
			It("lists all plugin data", func() {
				result := make(map[string][]clipr.Plugin)
				result["repo1"] = []clipr.Plugin{
					{
						Name:        "plugin1",
						Description: "none1",
					},
				}
				result["repo2"] = []clipr.Plugin{
					{
						Name:        "plugin2",
						Description: "none2",
					},
				}
				fakePluginRepo.GetPluginsReturns(result, []string{})

				err := callRepoPlugins()
				Expect(err).NotTo(HaveOccurred())

				Expect(ui.Outputs()).NotTo(ContainSubstrings([]string{"Logged errors:"}))
				Expect(ui.Outputs()).To(ContainSubstrings([]string{"repo1"}))
				Expect(ui.Outputs()).To(ContainSubstrings([]string{"plugin1"}))
				Expect(ui.Outputs()).To(ContainSubstrings([]string{"repo2"}))
				Expect(ui.Outputs()).To(ContainSubstrings([]string{"plugin2"}))
			})
		})

		Context("If errors are reported back from GetPlugins()", func() {
			It("informs user about the errors", func() {
				fakePluginRepo.GetPluginsReturns(nil, []string{
					"error from repo1",
					"error from repo2",
				})

				err := callRepoPlugins()
				Expect(err).NotTo(HaveOccurred())

				Expect(ui.Outputs()).To(ContainSubstrings([]string{"Logged errors:"}))
				Expect(ui.Outputs()).To(ContainSubstrings([]string{"error from repo1"}))
				Expect(ui.Outputs()).To(ContainSubstrings([]string{"error from repo2"}))
			})
		})
	})
})
