package commands_test

import (
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/pluginconfig"
	"github.com/cloudfoundry/cli/cf/configuration/pluginconfig/pluginconfigfakes"
	"github.com/cloudfoundry/cli/commandsloader"
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

	commandsloader.Load()

	var (
		ui                  *testterm.FakeUI
		requirementsFactory *testreq.FakeReqFactory
		config              *pluginconfigfakes.FakePluginConfiguration
		deps                commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.PluginConfig = config
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("help").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		requirementsFactory = &testreq.FakeReqFactory{}
		config = new(pluginconfigfakes.FakePluginConfiguration)
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("help", args, requirementsFactory, updateCommandDependency, false)
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
			m := make(map[string]pluginconfig.PluginMetadata)
			m["fakePlugin"] = pluginconfig.PluginMetadata{
				Commands: []plugin.Command{
					{
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
