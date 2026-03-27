package runner_test

import (
	"code.cloudfoundry.org/cli/v8/plugin/runner"
	"code.cloudfoundry.org/cli/v8/util/configv3"
	"code.cloudfoundry.org/cli/v8/util/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("PluginRunner", func() {
	var (
		config    *configv3.Config
		commandUI *ui.UI
		plugin    configv3.Plugin
		out       *Buffer
	)

	BeforeEach(func() {
		out = NewBuffer()
		commandUI = ui.NewTestUI(nil, out, NewBuffer())

		config = &configv3.Config{
			ConfigFile: configv3.JSONConfig{
				ConfigVersion: 3,
				Target:        "https://api.example.com",
				PluginRepositories: []configv3.PluginRepository{
					{
						Name: "CF-Community",
						URL:  "https://plugins.cloudfoundry.org",
					},
				},
			},
		}

		plugin = configv3.Plugin{
			Name:     "test-plugin",
			Location: "/path/to/test-plugin",
			Version: configv3.PluginVersion{
				Major: 1,
				Minor: 0,
				Build: 0,
			},
			Commands: []configv3.PluginCommand{
				{
					Name:     "test-command",
					Alias:    "tc",
					HelpText: "A test command",
					UsageDetails: configv3.PluginUsageDetails{
						Usage: "cf test-command [options]",
					},
				},
				{
					Name:     "another-command",
					HelpText: "Another test command",
					UsageDetails: configv3.PluginUsageDetails{
						Usage: "cf another-command",
					},
				},
			},
		}
	})

	Describe("NewPluginRunner", func() {
		It("creates a new PluginRunner instance", func() {
			pluginRunner := runner.NewPluginRunner(config, commandUI, plugin)
			Expect(pluginRunner).ToNot(BeNil())
		})

		It("accepts nil config", func() {
			pluginRunner := runner.NewPluginRunner(nil, commandUI, plugin)
			Expect(pluginRunner).ToNot(BeNil())
		})

		It("accepts nil UI", func() {
			pluginRunner := runner.NewPluginRunner(config, nil, plugin)
			Expect(pluginRunner).ToNot(BeNil())
		})

		It("accepts empty plugin", func() {
			emptyPlugin := configv3.Plugin{}
			pluginRunner := runner.NewPluginRunner(config, commandUI, emptyPlugin)
			Expect(pluginRunner).ToNot(BeNil())
		})
	})

	Describe("Run", func() {
		var pluginRunner runner.PluginRunner

		BeforeEach(func() {
			pluginRunner = runner.NewPluginRunner(config, commandUI, plugin)
		})

		Context("when the plugin binary does not exist", func() {
			It("returns ErrFailed", func() {
				err := pluginRunner.Run([]string{"test-command"})
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(runner.ErrFailed))
			})
		})

		Context("when the plugin binary path is empty", func() {
			BeforeEach(func() {
				plugin.Location = ""
				pluginRunner = runner.NewPluginRunner(config, commandUI, plugin)
			})

			It("returns an error", func() {
				err := pluginRunner.Run([]string{"test-command"})
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when args are empty", func() {
			It("can still execute (plugin determines behavior)", func() {
				// This will fail because the binary doesn't exist, but it tests
				// that empty args don't cause a panic
				err := pluginRunner.Run([]string{})
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(runner.ErrFailed))
			})
		})

		// Note: Integration tests with actual plugin binaries would require:
		// 1. A test plugin binary to be compiled
		// 2. Proper RPC communication setup
		// 3. Mock RPC server responses
		// These are typically done in integration test suites with fixture binaries
	})
})
