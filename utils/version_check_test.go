package utils_test

import (
	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/utils"

	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("NotifyUpdateIfNeeded", func() {

	var (
		ui     *testterm.FakeUI
		config core_config.ReadWriter
	)

	BeforeEach(func() {
		ui = new(testterm.FakeUI)
		config = testconfig.NewRepository()
	})

	It("Prints a notification to user if current version < min cli version", func() {
		config.SetMinCliVersion("6.0.0")
		config.SetMinRecommendedCliVersion("6.5.0")
		config.SetApiVersion("2.15.1")
		cf.Version = "5.0.0"
		NotifyUpdateIfNeeded(ui, config)

		Ω(ui.Outputs).To(ContainSubstrings([]string{"Cloud Foundry API version",
			"requires CLI version 6.0.0",
			"You are currently on version 5.0.0",
			"To upgrade your CLI, please visit: https://github.com/cloudfoundry/cli#downloads",
		}))
	})

	It("Doesn't print a notification to user if current version >= min cli version", func() {
		config.SetMinCliVersion("6.0.0")
		config.SetMinRecommendedCliVersion("6.5.0")
		config.SetApiVersion("2.15.1")
		cf.Version = "6.0.0"
		NotifyUpdateIfNeeded(ui, config)

		Ω(ui.Outputs).To(BeNil())
	})

})
