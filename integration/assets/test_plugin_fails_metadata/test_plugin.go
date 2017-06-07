package main

import (
	"os"

	"code.cloudfoundry.org/cli/plugin"
)

type TestPluginFailsMetadata struct{}

func (_ *TestPluginFailsMetadata) Run(cliConnection plugin.CliConnection, args []string) {
}

func (c *TestPluginFailsMetadata) GetMetadata() plugin.PluginMetadata {
	os.Exit(51)
	return plugin.PluginMetadata{
		Name: "CF-CLI-Panic-Integration-Test-Plugin",
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
			{Name: "freak-out"},
		},
	}
}

func uninstalling() {
}

func main() {
	plugin.Start(new(TestPluginFailsMetadata))
}
