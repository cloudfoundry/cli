package main

import (
	"fmt"
	"os"

	"code.cloudfoundry.org/cli/plugin"
)

type TestPluginWithCommandOverrides struct {
}

func (c *TestPluginWithCommandOverrides) Run(cliConnection plugin.CliConnection, args []string) {
	fmt.Println("How??? This should not even be allowed to run")
	os.Exit(1)
}

func (c *TestPluginWithCommandOverrides) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "CF-CLI-Command-Override-Integration-Test-Plugin",
		Version: plugin.VersionType{
			Major: 1,
			Minor: 2,
			Build: 4,
		},
		MinCliVersion: plugin.VersionType{
			Major: 5,
			Minor: 0,
			Build: 0,
		},
		Commands: []plugin.Command{
			{Name: "push", Alias: "p"},
		},
	}
}

func uninstalling() {
}

func main() {
	plugin.Start(new(TestPluginWithCommandOverrides))
}
