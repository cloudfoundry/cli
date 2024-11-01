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

type TestWithOrgs struct {
}

func (c *TestWithOrgs) Run(cliConnection plugin.CliConnection, args []string) {
	if args[0] == "orgs" {
		theOrgsCmd()
	}
}

func (c *TestWithOrgs) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "TestWithOrgs",
		Commands: []plugin.Command{
			{
				Name:     "orgs",
				HelpText: "",
			},
		},
	}
}

func theOrgsCmd() {
	fmt.Println("You called orgs in test_with_orgs")
}

func main() {
	plugin.Start(new(TestWithOrgs))
}
