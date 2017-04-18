package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"code.cloudfoundry.org/cli/plugin"
	"github.com/blang/semver"
)

var (
	pluginName     string
	commands       string
	commandHelps   string
	commandAliases string
	version        string
)

type ConfigurablePluginFailsUninstall struct{}

func (_ *ConfigurablePluginFailsUninstall) Run(cliConnection plugin.CliConnection, args []string) {
	fmt.Fprintf(os.Stderr, "I'm failing...I'm failing...\n")
	os.Exit(1)
}

func (_ *ConfigurablePluginFailsUninstall) GetMetadata() plugin.PluginMetadata {
	v1, _ := semver.Make(version)
	metadata := plugin.PluginMetadata{
		Name: pluginName,
		Version: plugin.VersionType{
			Major: int(v1.Major),
			Minor: int(v1.Minor),
			Build: int(v1.Patch),
		},
	}

	pluginCommandsList := strings.Split(commands, ",")
	pluginHelpsList := strings.Split(commandHelps, ",")
	pluginAliasesList := strings.Split(commandAliases, ",")
	for i, _ := range pluginCommandsList {
		metadata.Commands = append(metadata.Commands, plugin.Command{
			Alias:    pluginAliasesList[i],
			Name:     pluginCommandsList[i],
			HelpText: pluginHelpsList[i],
		})
	}

	return metadata
}

func uninstalling() {
	os.Remove(filepath.Join(os.TempDir(), "uninstall-test-file-for-test_1.exe"))
}

func main() {
	plugin.Start(new(ConfigurablePluginFailsUninstall))
}
