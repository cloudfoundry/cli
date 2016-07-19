/**
	* 1. Setup the server so cf can call it under main.
				e.g. `cf my-plugin` creates the callable server. now we can call the Run command
	* 2. Implement Run that is the actual code of the plugin!
	* 3. Return an error
**/

package main

import (
	"fmt"

	"code.cloudfoundry.org/cli/plugin"
)

type TestWithPushShortName struct {
}

func (c *TestWithPushShortName) Run(cliConnection plugin.CliConnection, args []string) {
	if args[0] == "p" {
		thePushCmd()
	}
}

func (c *TestWithPushShortName) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "TestWithPushShortName",
		Commands: []plugin.Command{
			{
				Name:     "p",
				HelpText: "plugin short name p",
			},
		},
	}
}

func thePushCmd() {
	fmt.Println("You called p within the plugin")
}

func main() {
	plugin.Start(new(TestWithPushShortName))
}
