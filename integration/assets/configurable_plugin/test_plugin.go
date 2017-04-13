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

type Test1 struct {
}

func (c *Test1) Run(cliConnection plugin.CliConnection, args []string) {
	fmt.Println("hello :)")
}

func (c *Test1) GetMetadata() plugin.PluginMetadata {
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
	plugin.Start(new(Test1))
}
