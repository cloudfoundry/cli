/**
	* 1. Setup the server so cf can call it under main.
				e.g. `cf my-plugin` creates the callable server. now we can call the Run command
	* 2. Implement Run that is the actual code of the plugin!
	* 3. Return an error
**/

package main

import (
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/plugin"
)

type MySay struct {
}

func (c *MySay) Run(cliConnection plugin.CliConnection, args []string) {
	if args[0] == "my-say" {
		if len(args) == 3 && args[2] == "--loud" {
			fmt.Println(strings.ToUpper(args[1]))
		}

		fmt.Println(args[1])
	}
}

func (c *MySay) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "MySay",
		Commands: []plugin.Command{
			{
				Name:     "my-say",
				HelpText: "Plugin to say things from the cli",
			},
		},
	}
}

func main() {
	plugin.Start(new(MySay))
}
