package models_test

import (
	"os"

	. "github.com/cloudfoundry-incubator/cli-plugin-repo/models"
	"github.com/cloudfoundry-incubator/cli-plugin-repo/test_helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Models", func() {
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

		It("populates the plugin model with raw data", func() {
			Ω(len(data)).To(Equal(2))
			Ω(data[0].Name).To(Equal("test1"))
			Ω(data[0].Binaries[0].Platform).To(Equal("osx"))
			Ω(data[1].Name).To(Equal("test2"))
			Ω(data[1].Binaries[1].Platform).To(Equal("linux32"))
		})

		It("turns optional string fields with nil value into empty string", func() {
			Ω(len(data)).To(Equal(2))
			Ω(data[0].Authors[0].Name).To(Equal("sample_name"))
			Ω(data[0].Company).To(Equal(""))
			Ω(data[0].Homepage).To(Equal(""))
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
