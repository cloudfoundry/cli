package helpers

import (
	"fmt"
	"os"
	"strings"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

type PluginCommand struct {
	Name  string
	Alias string
	Help  string
}

func InstallConfigurablePlugin(name string, version string, pluginCommands []PluginCommand) {
	path := BuildConfigurablePlugin("configurable_plugin", name, version, pluginCommands)
	Eventually(CF("install-plugin", "-f", path)).Should(Exit(0))
	Eventually(CFWithEnv(
		map[string]string{"CF_CLI_EXPERIMENTAL": "true"},
		"install-plugin", "-f", path)).Should(Exit(0))
}

func InstallConfigurablePluginFailsUninstall(name string, version string, pluginCommands []PluginCommand) {
	path := BuildConfigurablePlugin("configurable_plugin_fails_uninstall", name, version, pluginCommands)
	Eventually(CF("install-plugin", "-f", path)).Should(Exit(0))
}

func BuildConfigurablePlugin(pluginType string, name string, version string, pluginCommands []PluginCommand) string {
	commands := []string{}
	commandHelps := []string{}
	commandAliases := []string{}
	for _, command := range pluginCommands {
		commands = append(commands, command.Name)
		commandAliases = append(commandAliases, command.Alias)
		commandHelps = append(commandHelps, command.Help)
	}

	pluginPath, err := Build(fmt.Sprintf("code.cloudfoundry.org/cli/integration/assets/%s", pluginType),
		"-o",
		name,
		"-ldflags",
		fmt.Sprintf("-X main.pluginName=%s -X main.version=%s -X main.commands=%s -X main.commandHelps=%s -X main.commandAliases=%s",
			name,
			version,
			strings.Join(commands, ","),
			strings.Join(commandHelps, ","),
			strings.Join(commandAliases, ",")))
	Expect(err).ToNot(HaveOccurred())

	// gexec.Build builds the plugin with the name of the dir in the plugin path (configurable_plugin)
	// in case this function is called multiple times, the plugins need to be unique to be installed

	// also remove the .exe that gexec adds on Windows so the filename is always the
	// same in tests
	uniquePath := fmt.Sprintf("%s.%s", strings.TrimSuffix(pluginPath, ".exe"), name)
	err = os.Rename(pluginPath, uniquePath)
	Expect(err).ToNot(HaveOccurred())

	return uniquePath
}
