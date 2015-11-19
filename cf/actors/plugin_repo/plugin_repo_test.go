package plugin_repo_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	. "github.com/cloudfoundry/cli/cf/actors/plugin_repo"
	"github.com/cloudfoundry/cli/cf/models"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PluginRepo", func() {
	var (
		repoActor            PluginRepo
		testServer1CallCount int
		testServer2CallCount int
		testServer1          *httptest.Server
		testServer2          *httptest.Server
	)

	BeforeEach(func() {
		repoActor = NewPluginRepo()
	})

	Context("request data from all repos", func() {
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

		})

		AfterEach(func() {
			testServer1.Close()
			testServer2.Close()
		})

		It("make query to all repos listed in config.json", func() {
			repoActor.GetPlugins([]models.PluginRepo{
				models.PluginRepo{
					Name: "repo1",
					Url:  testServer1.URL,
				},
				models.PluginRepo{
					Name: "repo2",
					Url:  testServer2.URL,
				},
			})

			Ω(testServer1CallCount).To(Equal(1))
			Ω(testServer2CallCount).To(Equal(1))
		})

		It("lists each of the repos in config.json", func() {
			list, _ := repoActor.GetPlugins([]models.PluginRepo{
				models.PluginRepo{
					Name: "repo1",
					Url:  testServer1.URL,
				},
				models.PluginRepo{
					Name: "repo2",
					Url:  testServer2.URL,
				},
			})

			Ω(list["repo1"]).ToNot(BeNil())
			Ω(list["repo2"]).ToNot(BeNil())
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
							"version":"1.3.4",
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

			})

			AfterEach(func() {
				testServer1.Close()
			})

			It("lists the info for each plugin", func() {
				list, _ := repoActor.GetPlugins([]models.PluginRepo{
					models.PluginRepo{
						Name: "repo1",
						Url:  testServer1.URL,
					},
				})

				Ω(list["repo1"]).ToNot(BeNil())
				Ω(len(list["repo1"])).To(Equal(2))

				Ω(list["repo1"][0].Name).To(Equal("plugin1"))
				Ω(list["repo1"][0].Description).To(Equal("none"))
				Ω(list["repo1"][0].Version).To(Equal("1.3.4"))
				Ω(list["repo1"][0].Binaries[0].Platform).To(Equal("osx"))
				Ω(list["repo1"][1].Name).To(Equal("plugin2"))
				Ω(list["repo1"][1].Binaries[0].Platform).To(Equal("windows"))
			})

		})
	})

	Context("When data is invalid", func() {
		Context("json is invalid", func() {
			BeforeEach(func() {
				h1 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					fmt.Fprintln(w, `"plugins":[]}`)
				})
				testServer1 = httptest.NewServer(h1)
			})

			AfterEach(func() {
				testServer1.Close()
			})

			It("informs user of invalid json", func() {
				_, err := repoActor.GetPlugins([]models.PluginRepo{
					models.PluginRepo{
						Name: "repo1",
						Url:  testServer1.URL,
					},
				})

				Ω(err).To(ContainSubstrings(
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
			})

			AfterEach(func() {
				testServer1.Close()
			})

			It("informs user of invalid repo data", func() {
				_, err := repoActor.GetPlugins([]models.PluginRepo{
					models.PluginRepo{
						Name: "repo1",
						Url:  testServer1.URL,
					},
				})

				Ω(err).To(ContainSubstrings(
					[]string{"Invalid data", "plugin data does not exist"},
				))
			})

		})
	})
})
