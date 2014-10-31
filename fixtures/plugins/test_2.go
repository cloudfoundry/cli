/**
	* 1. Setup the server so cf can call it under main.
				e.g. `cf my-plugin` creates the callable server. now we can call the Run command
	* 2. Implement Run that is the actual code of the plugin!
	* 3. Return an error
**/

package main

import (
	"fmt"

	"github.com/cloudfoundry/cli/plugin"
)

type Test2 struct{}

func (c *Test2) Run(args []string) {
	if args[0] == "test_2_cmd1" {
		theFirstCmd()
	} else if args[0] == "test_2_cmd2" {
		theSecondCmd()
	}
}

func (c *Test2) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "Test2",
		Commands: []plugin.Command{
			{
				Name:     "test_2_cmd1",
				HelpText: "help text for test_2_cmd1",
			},
			{
				Name:     "test_2_cmd2",
				HelpText: "help text for test_2_cmd2",
			},
		},
	}
}

func theFirstCmd() {
	fmt.Println("You called cmd1 in test_2")
}

func theSecondCmd() {
	fmt.Println("You called cmd2 in test_2")
}

func main() {
	plugin.Start(new(Test2))
}
