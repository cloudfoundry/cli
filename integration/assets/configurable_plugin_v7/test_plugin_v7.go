package main

import (
	"fmt"
	"os"
	"strings"

	plugin "code.cloudfoundry.org/cli/plugin/v7"
	"github.com/blang/semver"
)

var (
	pluginName     string
	commands       string
	commandHelps   string
	commandAliases string
	version        string
)

type ConfigurablePlugin struct {
}

func (_ *ConfigurablePlugin) Run(cliConnection plugin.CliConnection, args []string) {
	fmt.Printf("%s\n", strings.Join(os.Args, " "))
}

func (_ *ConfigurablePlugin) GetMetadata() plugin.PluginMetadata {
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

func main() {
	plugin.Start(new(ConfigurablePlugin))
}
