package plugin_repo_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"

	. "github.com/cloudfoundry/cli/cf/commands/plugin_repo"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"

	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("repo-plugins", func() {
	var (
		ui                  *testterm.FakeUI
		config              core_config.ReadWriter
		requirementsFactory *testreq.FakeReqFactory
		testServer1         *httptest.Server
		testServer2         *httptest.Server
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		requirementsFactory = &testreq.FakeReqFactory{}
		config = testconfig.NewRepositoryWithDefaults()
	})

	var callRepoPlugins = func(args ...string) bool {
		cmd := NewRepoPlugins(ui, config)
		return testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	Context("If repo name is provided by '-r'", func() {
		Context("request data from named repos", func() {
			var (
				testServer1CallCount int
				testServer2CallCount int
			)

			BeforeEach(func() {
				testServer1CallCount = 0
				h1 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					testServer1CallCount++
					fmt.Fprintln(w, `{"plugins":[]}`)
				})
				testServer1 = httptest.NewServer(h1)

				testServer2CallCount = 0
				h2 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					testServer2CallCount++
					fmt.Fprintln(w, `{"plugins":[]}`)
				})
				testServer2 = httptest.NewServer(h2)

				config.SetPluginRepo(models.PluginRepo{
					Name: "repo1",
					Url:  testServer1.URL,
				})

				config.SetPluginRepo(models.PluginRepo{
					Name: "repo2",
					Url:  testServer2.URL,
				})
			})

			AfterEach(func() {
				testServer1.Close()
				testServer2.Close()
			})

			It("make query to just the repo named", func() {
				callRepoPlugins("-r", "repo2")

				Ω(testServer1CallCount).To(Equal(0))
				Ω(testServer2CallCount).To(Equal(1))
			})

			It("informs user if requested repo is not found", func() {
				callRepoPlugins("-r", "repo_not_there")

				Ω(testServer1CallCount).To(Equal(0))
				Ω(testServer2CallCount).To(Equal(0))
				Ω(ui.Outputs).To(ContainSubstrings([]string{"repo_not_there", "does not exist as an available plugin repo"}))
				Ω(ui.Outputs).To(ContainSubstrings([]string{"Tip: use `add-plugin-repo` command to add repos."}))
			})

		})
	})

	Context("If no repo name is provided", func() {
		Context("request data from all repos", func() {
			var (
				testServer1CallCount int
				testServer2CallCount int
			)

			BeforeEach(func() {
				testServer1CallCount = 0
				h1 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					testServer1CallCount++
					fmt.Fprintln(w, `{"plugins":[]}`)
				})
				testServer1 = httptest.NewServer(h1)

				testServer2CallCount = 0
				h2 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					testServer2CallCount++
					fmt.Fprintln(w, `{"plugins":[]}`)
				})
				testServer2 = httptest.NewServer(h2)

				config.SetPluginRepo(models.PluginRepo{
					Name: "repo1",
					Url:  testServer1.URL,
				})

				config.SetPluginRepo(models.PluginRepo{
					Name: "repo2",
					Url:  testServer2.URL,
				})
			})

			AfterEach(func() {
				testServer1.Close()
				testServer2.Close()
			})

			It("make query to all repos listed in config.json", func() {
				callRepoPlugins()

				Ω(testServer1CallCount).To(Equal(1))
				Ω(testServer2CallCount).To(Equal(1))
			})

			It("lists each of the repos in config.json", func() {
				callRepoPlugins()

				Ω(ui.Outputs).To(ContainSubstrings(
					[]string{"repo1"},
					[]string{"repo2"},
				))
			})

		})

	})

	Context("Getting data from repos", func() {
		Context("When data is valid", func() {
			BeforeEach(func() {
				h1 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					fmt.Fprintln(w, `{"plugins":[
						{
							"name":"plugin1",
							"description":"none",
							"version":"4",
							"binaries":[
								{
									"platform":"osx",
									"url":"https://github.com/simonleung8/cli-plugin-echo/raw/master/bin/osx/echo",
									"checksum":"2a087d5cddcfb057fbda91e611c33f46"
								}
							]
						},
						{
							"name":"plugin2",
							"binaries":[
								{
									"platform":"windows",
									"url":"http://going.no.where",
									"checksum":"abcdefg"
								}
							]
						}]
					}`)
				})
				testServer1 = httptest.NewServer(h1)

				config.SetPluginRepo(models.PluginRepo{
					Name: "repo1",
					Url:  testServer1.URL,
				})
			})

			AfterEach(func() {
				testServer1.Close()
			})

			It("lists the info for each plugin", func() {
				callRepoPlugins()

				Ω(ui.Outputs).To(ContainSubstrings(
					[]string{"repo1"},
					[]string{"plugin1", "4", "none"},
					[]string{"plugin2"},
				))
			})

		})

		Context("When data is invalid", func() {
			Context("json is invalid", func() {
				BeforeEach(func() {
					h1 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						fmt.Fprintln(w, `"plugins":[]}`)
					})
					testServer1 = httptest.NewServer(h1)

					config.SetPluginRepo(models.PluginRepo{
						Name: "repo1",
						Url:  testServer1.URL,
					})
				})

				AfterEach(func() {
					testServer1.Close()
				})

				It("informs user of invalid json", func() {
					callRepoPlugins()

					Ω(ui.Outputs).To(ContainSubstrings(
						[]string{"Invalid json data"},
					))
				})

			})

			Context("when data is valid json, but not valid plugin repo data", func() {
				BeforeEach(func() {
					h1 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						fmt.Fprintln(w, `{"bad_plugin_tag":[]}`)
					})
					testServer1 = httptest.NewServer(h1)

					config.SetPluginRepo(models.PluginRepo{
						Name: "repo1",
						Url:  testServer1.URL,
					})
				})

				AfterEach(func() {
					testServer1.Close()
				})

				It("informs user of invalid repo data", func() {
					callRepoPlugins()

					Ω(ui.Outputs).To(ContainSubstrings(
						[]string{"Invalid data", "plugin data does not exist"},
					))
				})

			})
		})
	})

})
