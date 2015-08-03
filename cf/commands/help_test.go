package commands_test

import (
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/plugin_config"
	testconfig "github.com/cloudfoundry/cli/cf/configuration/plugin_config/fakes"
	"github.com/cloudfoundry/cli/commands_loader"
	"github.com/cloudfoundry/cli/plugin"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	io_helpers "github.com/cloudfoundry/cli/testhelpers/io"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Help", func() {

	commands_loader.Load()

	var (
		ui                  *testterm.FakeUI
		requirementsFactory *testreq.FakeReqFactory
		config              *testconfig.FakePluginConfiguration
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.PluginConfig = config
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("help").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		requirementsFactory = &testreq.FakeReqFactory{}
		config = &testconfig.FakePluginConfiguration{}
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCliCommand("help", args, requirementsFactory, updateCommandDependency, false)
	}

	Context("when no argument is provided", func() {
		It("prints the main help menu of the 'cf' app", func() {
			outputs := io_helpers.CaptureOutput(func() { runCommand() })

			Eventually(outputs).Should(ContainSubstrings([]string{"A command line tool to interact with Cloud Foundry"}))
			Eventually(outputs).Should(ContainSubstrings([]string{"CF_TRACE=true"}))
		})
	})

	Context("when a command name is provided as an argument", func() {
		Context("When the command exists", func() {
			It("prints the usage help for the command", func() {
				runCommand("target")

				Eventually(ui.Outputs).Should(ContainSubstrings([]string{"target - Set or view the targeted org or space"}))
			})
		})

		Context("When the command exists", func() {
			It("prints the usage help for the command", func() {
				runCommand("bad-command")

				Eventually(ui.Outputs).Should(ContainSubstrings([]string{"'bad-command' is not a registered command. See 'cf help'"}))
			})
		})
	})

	Context("when a command provided is a plugin command", func() {
		BeforeEach(func() {
			m := make(map[string]plugin_config.PluginMetadata)
			m["fakePlugin"] = plugin_config.PluginMetadata{
				Commands: []plugin.Command{
					plugin.Command{
						Name:     "fakePluginCmd1",
						Alias:    "fpc1",
						HelpText: "help text here",
						UsageDetails: plugin.Usage{
							Usage: "Usage for fpc1",
							Options: map[string]string{
								"f": "test flag",
							},
						},
					},
				},
			}

			config.PluginsReturns(m)
		})

		Context("command is a plugin command name", func() {
			It("prints the usage help for the command", func() {
				runCommand("fakePluginCmd1")

				Eventually(ui.Outputs).Should(ContainSubstrings([]string{"fakePluginCmd1", "help text here"}))
				Eventually(ui.Outputs).Should(ContainSubstrings([]string{"ALIAS"}))
				Eventually(ui.Outputs).Should(ContainSubstrings([]string{"fpc1"}))
				Eventually(ui.Outputs).Should(ContainSubstrings([]string{"USAGE"}))
				Eventually(ui.Outputs).Should(ContainSubstrings([]string{"Usage for fpc1"}))
				Eventually(ui.Outputs).Should(ContainSubstrings([]string{"OPTIONS"}))
				Eventually(ui.Outputs).Should(ContainSubstrings([]string{"-f", "test flag"}))
			})
		})

		Context("command is a plugin command alias", func() {
			It("prints the usage help for the command alias", func() {
				runCommand("fpc1")

				Eventually(ui.Outputs).Should(ContainSubstrings([]string{"fakePluginCmd1", "help text here"}))
				Eventually(ui.Outputs).Should(ContainSubstrings([]string{"ALIAS"}))
				Eventually(ui.Outputs).Should(ContainSubstrings([]string{"fpc1"}))
				Eventually(ui.Outputs).Should(ContainSubstrings([]string{"USAGE"}))
				Eventually(ui.Outputs).Should(ContainSubstrings([]string{"Usage for fpc1"}))
				Eventually(ui.Outputs).Should(ContainSubstrings([]string{"OPTIONS"}))
				Eventually(ui.Outputs).Should(ContainSubstrings([]string{"-f", "test flag"}))
			})
		})

	})
})
