package plugin_repo_test

import (
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/commands/plugin_repo"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("delte-plugin-repo", func() {
	var (
		ui                  *testterm.FakeUI
		config              core_config.ReadWriter
		requirementsFactory *testreq.FakeReqFactory
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		requirementsFactory = &testreq.FakeReqFactory{}
		config = testconfig.NewRepositoryWithDefaults()
	})

	var callRemovePluginRepo = func(args ...string) bool {
		cmd := NewRemovePluginRepo(ui, config)
		return testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	Context("When repo name is valid", func() {
		BeforeEach(func() {
			config.SetPluginRepo(models.PluginRepo{
				Name: "repo1",
				Url:  "http://someserver1.com:1234",
			})

			config.SetPluginRepo(models.PluginRepo{
				Name: "repo2",
				Url:  "http://server2.org:8080",
			})
		})

		It("deletes the repo from the config", func() {
			callRemovePluginRepo("repo1")
			Ω(len(config.PluginRepos())).To(Equal(1))
			Ω(config.PluginRepos()[0].Name).To(Equal("repo2"))
			Ω(config.PluginRepos()[0].Url).To(Equal("http://server2.org:8080"))
		})
	})

	Context("When named repo doesn't exist", func() {
		BeforeEach(func() {
			config.SetPluginRepo(models.PluginRepo{
				Name: "repo1",
				Url:  "http://someserver1.com:1234",
			})

			config.SetPluginRepo(models.PluginRepo{
				Name: "repo2",
				Url:  "http://server2.org:8080",
			})
		})

		It("doesn't change the config the config", func() {
			callRemovePluginRepo("fake-repo")

			Ω(len(config.PluginRepos())).To(Equal(2))
			Ω(config.PluginRepos()[0].Name).To(Equal("repo1"))
			Ω(config.PluginRepos()[0].Url).To(Equal("http://someserver1.com:1234"))
			Ω(ui.Outputs).Should(ContainSubstrings([]string{"fake-repo", "does not exist as a repo"}))
		})
	})
})
