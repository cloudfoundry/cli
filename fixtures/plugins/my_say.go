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

	"github.com/cloudfoundry/cli/plugin"
)

type MySay struct {
	stringForMySay string
}

func (c *MySay) Run(args []string, reply *bool) error {
	if args[0] == "my-say" {
		if len(args) == 3 && args[2] == "--loud" {
			fmt.Println(strings.ToUpper(args[1]))

			*reply = true
			return nil
		}

		fmt.Println(args[1])
	}
	*reply = true
	return nil
}

func (c *MySay) GetCommands() []plugin.Command {
	return []plugin.Command{
		{
			Name:     "my-say",
			HelpText: "Plugin to say things from the cli",
		},
	}
}

func main() {
	plugin.Start(new(MySay))
}
