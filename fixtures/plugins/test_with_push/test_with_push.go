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

type TestWithPush struct {
}

func (c *TestWithPush) Run(cliConnection plugin.CliConnection, args []string) {
	if args[0] == "push" {
		thePushCmd()
	}
}

func (c *TestWithPush) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "TestWithPush",
		Commands: []plugin.Command{
			{
				Name:     "push",
				HelpText: "push text for test_with_push",
			},
		},
	}
}

func thePushCmd() {
	fmt.Println("You called push in test_with_push")
}

func main() {
	plugin.Start(new(TestWithPush))
}
