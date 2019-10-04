// +build V7

package main

import (
	"fmt"
	"os"

	plugin "code.cloudfoundry.org/cli/plugin/v7"
)

type AppPlugin struct{}

func (c *AppPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	if args[0] == "v7-app-plugin" {
		app, err := cliConnection.GetApp("dora")
		if err != nil {
			panic(err)
		}

		fmt.Fprintf(os.Stdout, "%+v\n", app)
	}
}

func (c *AppPlugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "v7-app-plugin",
		Version: plugin.VersionType{
			Major: 1,
			Minor: 0,
			Build: 0,
		},
		MinCliVersion: plugin.VersionType{
			Major: 6,
			Minor: 7,
			Build: 0,
		},
		Commands: []plugin.Command{
			{
				Name:     "v7-app-plugin",
				HelpText: "v7-app-plugin",
			},
		},
	}
}

func main() {
	plugin.Start(new(AppPlugin))
}
