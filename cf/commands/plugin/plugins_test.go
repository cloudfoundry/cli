package plugin_test

import (
	"net/rpc"

	"github.com/cloudfoundry/cli/cf/commandregistry"
	plugincmd "github.com/cloudfoundry/cli/cf/commands/plugin"
	"github.com/cloudfoundry/cli/cf/configuration/pluginconfig"
	"github.com/cloudfoundry/cli/cf/configuration/pluginconfig/pluginconfigfakes"
	"github.com/cloudfoundry/cli/cf/flags"
	"github.com/cloudfoundry/cli/cf/requirements/requirementsfakes"
	"github.com/cloudfoundry/cli/plugin"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Plugins", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *requirementsfakes.FakeFactory
		config              *pluginconfigfakes.FakePluginConfiguration
		deps                commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.PluginConfig = config
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("plugins").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		requirementsFactory = new(requirementsfakes.FakeFactory)
		config = new(pluginconfigfakes.FakePluginConfiguration)

		rpc.DefaultServer = rpc.NewServer()
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("plugins", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	Context("If --checksum flag is provided", func() {
		It("computes and prints the sha1 checksum of the binary", func() {
			config.PluginsReturns(map[string]pluginconfig.PluginMetadata{
				"Test1": {
					Location: "../../../fixtures/plugins/test_1.go",
					Version:  plugin.VersionType{Major: 1, Minor: 2, Build: 3},
					Commands: []plugin.Command{
						{Name: "test_1_cmd1", HelpText: "help text for test_1_cmd1"},
					},
				},
			})

			runCommand("--checksum")

			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Plugin Name", "Version", "sha1", "Command Help"},
			))
		})
	})

	Context("when arguments are provided", func() {
		var cmd commandregistry.Command
		var flagContext flags.FlagContext

		BeforeEach(func() {
			cmd = &plugincmd.Plugins{}
			cmd.SetDependency(deps, false)
			flagContext = flags.NewFlagContext(cmd.MetaData().Flags)
		})

		It("should fail with usage", func() {
			flagContext.Parse("blahblah")

			reqs := cmd.Requirements(requirementsFactory, flagContext)

			err := testcmd.RunRequirements(reqs)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Incorrect Usage"))
			Expect(err.Error()).To(ContainSubstring("No argument required"))
		})
	})

	It("returns a list of available methods of a plugin", func() {
		config.PluginsReturns(map[string]pluginconfig.PluginMetadata{
			"Test1": {
				Location: "path/to/plugin",
				Commands: []plugin.Command{
					{Name: "test_1_cmd1", HelpText: "help text for test_1_cmd1"},
					{Name: "test_1_cmd2", HelpText: "help text for test_1_cmd2"},
				},
			},
		})

		runCommand()

		Expect(ui.Outputs()).To(ContainSubstrings(
			[]string{"Listing Installed Plugins..."},
			[]string{"OK"},
			[]string{"Plugin Name", "Command Name", "Command Help"},
			[]string{"Test1", "test_1_cmd1", "help text for test_1_cmd1"},
			[]string{"Test1", "test_1_cmd2", "help text for test_1_cmd2"},
		))
	})

	It("lists the name of the command, it's alias and version", func() {
		config.PluginsReturns(map[string]pluginconfig.PluginMetadata{
			"Test1": {
				Location: "path/to/plugin",
				Version:  plugin.VersionType{Major: 1, Minor: 2, Build: 3},
				Commands: []plugin.Command{
					{Name: "test_1_cmd1", Alias: "test_1_cmd1_alias", HelpText: "help text for test_1_cmd1"},
					{Name: "test_1_cmd2", Alias: "test_1_cmd2_alias", HelpText: "help text for test_1_cmd2"},
				},
			},
		})

		runCommand()

		Expect(ui.Outputs()).To(ContainSubstrings(
			[]string{"Test1", "test_1_cmd1", "1.2.3", ", test_1_cmd1_alias", "help text for test_1_cmd1"},
			[]string{"Test1", "test_1_cmd2", "1.2.3", ", test_1_cmd2_alias", "help text for test_1_cmd2"},
		))
	})

	It("lists 'N/A' as version when plugin does not provide a version", func() {
		config.PluginsReturns(map[string]pluginconfig.PluginMetadata{
			"Test1": {
				Location: "path/to/plugin",
				Commands: []plugin.Command{
					{Name: "test_1_cmd1", Alias: "test_1_cmd1_alias", HelpText: "help text for test_1_cmd1"},
				},
			},
		})

		runCommand()

		Expect(ui.Outputs()).To(ContainSubstrings(
			[]string{"Test1", "test_1_cmd1", "N/A", ", test_1_cmd1_alias", "help text for test_1_cmd1"},
		))
	})

	It("does not list the plugin when it provides no available commands", func() {
		config.PluginsReturns(map[string]pluginconfig.PluginMetadata{
			"EmptyPlugin": {Location: "../../../fixtures/plugins/empty_plugin.exe"},
		})

		runCommand()
		Expect(ui.Outputs()).NotTo(ContainSubstrings(
			[]string{"EmptyPlugin"},
		))
	})

	It("list multiple plugins and their associated commands", func() {
		config.PluginsReturns(map[string]pluginconfig.PluginMetadata{
			"Test1": {Location: "path/to/plugin1", Commands: []plugin.Command{{Name: "test_1_cmd1", HelpText: "help text for test_1_cmd1"}}},
			"Test2": {Location: "path/to/plugin2", Commands: []plugin.Command{{Name: "test_2_cmd1", HelpText: "help text for test_2_cmd1"}}},
		})

		runCommand()
		Expect(ui.Outputs()).To(ContainSubstrings(
			[]string{"Test1", "test_1_cmd1", "help text for test_1_cmd1"},
			[]string{"Test2", "test_2_cmd1", "help text for test_2_cmd1"},
		))
	})
})
