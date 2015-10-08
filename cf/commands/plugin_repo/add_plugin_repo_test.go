package plugin_repo_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("add-plugin-repo", func() {
	var (
		ui                  *testterm.FakeUI
		config              core_config.Repository
		requirementsFactory *testreq.FakeReqFactory
		testServer          *httptest.Server
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.Config = config
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("add-plugin-repo").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		requirementsFactory = &testreq.FakeReqFactory{}
		config = testconfig.NewRepositoryWithDefaults()
	})

	var callAddPluginRepo = func(args []string) bool {
		return testcmd.RunCliCommand("add-plugin-repo", args, requirementsFactory, updateCommandDependency, false)
	}

	Context("When repo server is valid", func() {
		BeforeEach(func() {

			h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, `{"plugins":[
				{
					"name":"echo",
					"description":"none",
					"version":"4",
					"binaries":[
						{
							"platform":"osx",
							"url":"https://github.com/simonleung8/cli-plugin-echo/raw/master/bin/osx/echo",
							"checksum":"2a087d5cddcfb057fbda91e611c33f46"
						}
					]
				}]
			}`)
			})
			testServer = httptest.NewServer(h)
		})

		AfterEach(func() {
			testServer.Close()
		})

		It("saves the repo url into config", func() {
			callAddPluginRepo([]string{"repo", testServer.URL})

			Ω(config.PluginRepos()[0].Name).To(Equal("repo"))
			Ω(config.PluginRepos()[0].Url).To(Equal(testServer.URL))
		})
	})

	Context("repo name already existing", func() {
		BeforeEach(func() {
			config.SetPluginRepo(models.PluginRepo{Name: "repo", Url: "http://repo.com"})
		})

		It("informs user of the already existing repo", func() {

			callAddPluginRepo([]string{"repo", "http://repo2.com"})

			Ω(ui.Outputs).To(ContainSubstrings(
				[]string{"Plugin repo named \"repo\"", " already exists"},
			))
		})
	})

	Context("repo address already existing", func() {
		BeforeEach(func() {
			config.SetPluginRepo(models.PluginRepo{Name: "repo1", Url: "http://repo.com"})
		})

		It("informs user of the already existing repo", func() {

			callAddPluginRepo([]string{"repo2", "http://repo.com"})

			Ω(ui.Outputs).To(ContainSubstrings(
				[]string{"http://repo.com (repo1)", " already exists."},
			))
		})
	})

	Context("When repo server is not valid", func() {

		Context("server url is invalid", func() {
			It("informs user of invalid url which does not has prefix http", func() {

				callAddPluginRepo([]string{"repo", "msn.com"})

				Ω(ui.Outputs).To(ContainSubstrings(
					[]string{"msn.com", "is not a valid url"},
				))
			})
		})

		Context("server does not has a '/list' endpoint", func() {
			It("informs user of invalid repo server", func() {

				callAddPluginRepo([]string{"repo", "https://google.com"})

				Ω(ui.Outputs).To(ContainSubstrings(
					[]string{"https://google.com/list", "is not responding."},
				))
			})
		})

		Context("server responses with invalid json", func() {
			BeforeEach(func() {

				h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					fmt.Fprintln(w, `"plugins":[]}`)
				})
				testServer = httptest.NewServer(h)
			})

			AfterEach(func() {
				testServer.Close()
			})

			It("informs user of invalid repo server", func() {
				callAddPluginRepo([]string{"repo", testServer.URL})

				Ω(ui.Outputs).To(ContainSubstrings(
					[]string{"Error processing data from server"},
				))
			})
		})

		Context("server responses with json without 'plugins' object", func() {
			BeforeEach(func() {

				h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					fmt.Fprintln(w, `{"bad_plugins":[
						{
							"name": "plugin1",
							"description": "none"
						}
					]}`)
				})
				testServer = httptest.NewServer(h)
			})

			AfterEach(func() {
				testServer.Close()
			})

			It("informs user of invalid repo server", func() {
				callAddPluginRepo([]string{"repo", testServer.URL})

				Ω(ui.Outputs).To(ContainSubstrings(
					[]string{"\"Plugins\" object not found in the responded data"},
				))
			})
		})

	})

})
