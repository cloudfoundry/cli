package pluginrepo_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

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

var _ = Describe("add-plugin-repo", func() {
	var (
		ui                  *testterm.FakeUI
		config              coreconfig.Repository
		requirementsFactory *requirementsfakes.FakeFactory
		testServer          *httptest.Server
		deps                commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.Config = config
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("add-plugin-repo").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		requirementsFactory = new(requirementsfakes.FakeFactory)
		config = testconfig.NewRepositoryWithDefaults()
	})

	var callAddPluginRepo = func(args []string) bool {
		return testcmd.RunCLICommand("add-plugin-repo", args, requirementsFactory, updateCommandDependency, false, ui)
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

			Expect(config.PluginRepos()[0].Name).To(Equal("repo"))
			Expect(config.PluginRepos()[0].URL).To(Equal(testServer.URL))
		})
	})

	Context("repo name already existing", func() {
		BeforeEach(func() {
			config.SetPluginRepo(models.PluginRepo{Name: "repo", URL: "http://repo.com"})
		})

		It("informs user of the already existing repo", func() {

			callAddPluginRepo([]string{"repo", "http://repo2.com"})

			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Plugin repo named \"repo\"", " already exists"},
			))
		})
	})

	Context("repo address already existing", func() {
		BeforeEach(func() {
			config.SetPluginRepo(models.PluginRepo{Name: "repo1", URL: "http://repo.com"})
		})

		It("informs user of the already existing repo", func() {

			callAddPluginRepo([]string{"repo2", "http://repo.com"})

			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"http://repo.com (repo1)", " already exists."},
			))
		})
	})

	Context("When repo server is not valid", func() {

		Context("server url is invalid", func() {
			It("informs user of invalid url which does not has prefix http", func() {

				callAddPluginRepo([]string{"repo", "msn.com"})

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"msn.com", "is not a valid url"},
				))
			})

			It("should not contain the tip", func() {
				callAddPluginRepo([]string{"repo", "msn.com"})

				Expect(ui.Outputs()).NotTo(ContainSubstrings(
					[]string{"TIP: If you are behind a firewall and require an HTTP proxy, verify the https_proxy environment variable is correctly set. Else, check your network connection."},
				))
			})
		})

		Context("server does not has a '/list' endpoint", func() {
			It("informs user of invalid repo server", func() {

				callAddPluginRepo([]string{"repo", "https://google.com"})

				Expect(ui.Outputs()).To(ContainSubstrings(
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

				Expect(ui.Outputs()).To(ContainSubstrings(
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

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"\"Plugins\" object not found in the responded data"},
				))
			})
		})

		Context("When connection could not be established", func() {
			It("prints a tip", func() {
				callAddPluginRepo([]string{"repo", "https://broccoli.nonexistanttld:"})

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"TIP: If you are behind a firewall and require an HTTP proxy, verify the https_proxy environment variable is correctly set. Else, check your network connection."},
				))
			})
		})

		Context("server responds with an http error", func() {
			BeforeEach(func() {
				h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				})
				testServer = httptest.NewServer(h)
			})

			AfterEach(func() {
				testServer.Close()
			})

			It("does not print a tip", func() {
				callAddPluginRepo([]string{"repo", testServer.URL})

				Expect(ui.Outputs()).NotTo(ContainSubstrings(
					[]string{"TIP: If you are behind a firewall and require an HTTP proxy, verify the https_proxy environment variable is correctly set. Else, check your network connection."},
				))
			})
		})
	})
})
