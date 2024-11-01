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

type TestWithHelp struct {
}

func (c *TestWithHelp) Run(cliConnection plugin.CliConnection, args []string) {
	if args[0] == "help" {
		theHelpCmd()
	}
}

func (c *TestWithHelp) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "TestWithHelp",
		Commands: []plugin.Command{
			{
				Name:     "help",
				HelpText: "help text for test_with_help",
			},
		},
	}
}

func theHelpCmd() {
	fmt.Println("You called help in test_with_help")
}

func main() {
	plugin.Start(new(TestWithHelp))
}
