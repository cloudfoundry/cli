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

type TestWithOrgsShortName struct {
}

func (c *TestWithOrgsShortName) Run(cliConnection plugin.CliConnection, args []string) {
	if args[0] == "o" {
		theOrgsCmd()
	}
}

func (c *TestWithOrgsShortName) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "TestWithOrgsShortName",
		Commands: []plugin.Command{
			{
				Name:     "o",
				HelpText: "",
			},
		},
	}
}

func theOrgsCmd() {
	fmt.Println("You called o in test_with_orgs_short_name")
}

func main() {
	plugin.Start(new(TestWithOrgsShortName))
}
