package models_test

import (
	"os"
	"sort"
	"time"

	. "github.com/cloudfoundry-incubator/cli-plugin-repo/models"
	"github.com/cloudfoundry-incubator/cli-plugin-repo/test_helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("package models", func() {
	Context("PluginModel", func() {
		var (
			parsedYaml  interface{}
			pluginModel PluginModel
			data        []Plugin
		)

		Context("When raw data is valid", func() {
			BeforeEach(func() {
				parsedYaml = map[interface{}]interface{}{
					"plugins": []interface{}{
						map[interface{}]interface{}{
							"name":        "test1",
							"description": "n/a",
							"updated":     time.Date(2014, time.November, 10, 23, 0, 0, 0, time.UTC),
							"authors": []interface{}{
								map[interface{}]interface{}{
									"name":     "sample_name",
									"homepage": "",
									"contact":  "cant.find.me@mars.com",
								},
							},
							"binaries": []interface{}{
								map[interface{}]interface{}{
									"platform": "osx",
									"url":      "example.com/plugin",
									"checksum": "abcdefg",
								},
							},
						},
						map[interface{}]interface{}{
							"name":        "test2",
							"description": "n/a",
							"updated":     time.Date(2014, time.December, 01, 23, 0, 0, 0, time.UTC),
							"binaries": []interface{}{
								map[interface{}]interface{}{
									"platform": "windows",
									"url":      "example.com/plugin",
									"checksum": "abcdefg",
								},
								map[interface{}]interface{}{
									"platform": "linux32",
									"url":      "example.com/plugin",
									"checksum": "abcdefg",
								},
							},
						},
					},
				}

				pluginModel = NewPlugins(os.Stdout)
				data = pluginModel.PopulateModel(parsedYaml)
			})

			It("populates the plugin model with raw data sorted by last updated date", func() {
				Ω(len(data)).To(Equal(2))
				Ω(data[0].Name).To(Equal("test2"))
				Ω(data[0].Binaries[1].Platform).To(Equal("linux32"))
				Ω(data[1].Name).To(Equal("test1"))
				Ω(data[1].Binaries[0].Platform).To(Equal("osx"))
				Ω(data[0].Updated.After(data[1].Updated)).To(BeTrue())
			})

			It("turns optional string fields with nil value into empty string", func() {
				Ω(len(data)).To(Equal(2))
				Ω(data[1].Authors[0].Name).To(Equal("sample_name"))
				Ω(data[1].Company).To(Equal(""))
				Ω(data[1].Homepage).To(Equal(""))
			})
		})

		Context("When raw data contains unknown field", func() {
			var (
				logger *test_helpers.TestLogger
			)

			BeforeEach(func() {
				parsedYaml = map[interface{}]interface{}{
					"plugins": []interface{}{
						map[interface{}]interface{}{
							"name":          "test1",
							"description":   "n/a",
							"unknown_field": "123",
						},
					},
				}

				logger = test_helpers.NewTestLogger()
				pluginModel = NewPlugins(logger)
				data = pluginModel.PopulateModel(parsedYaml)
			})

			It("logs error to terminal", func() {
				Ω(len(data)).To(Equal(1))
				Ω(logger.ContainsSubstring([]string{"unexpected field", "unknown_field"})).To(Equal(true))
			})
		})

	})

	Context("PluginsJson", func() {
		var pluginJson PluginsJson

		BeforeEach(func() {
			pluginJson.Plugins = []Plugin{
				Plugin{
					Name:    "plugin1",
					Updated: time.Date(2014, time.November, 10, 23, 0, 0, 0, time.UTC),
				},
				Plugin{
					Name:    "plugin2",
					Updated: time.Date(2015, time.January, 10, 23, 0, 0, 0, time.UTC),
				},
				Plugin{
					Name:    "plugin3",
					Updated: time.Date(2015, time.January, 15, 23, 0, 0, 0, time.UTC),
				},
			}
		})

		Context("Sort", func() {
			It("sorts the array by last 'Updated' date", func() {
				sort.Sort(pluginJson)
				Ω(pluginJson.Plugins[0].Name).To(Equal("plugin3"))
				Ω(pluginJson.Plugins[1].Name).To(Equal("plugin2"))
				Ω(pluginJson.Plugins[2].Name).To(Equal("plugin1"))
			})
		})
	})
})
