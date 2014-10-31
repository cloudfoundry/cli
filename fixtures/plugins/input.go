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

type Input struct {
}

func (c *Input) Run(args []string) {
	if args[0] == "input" {
		var Echo string
		fmt.Scanf("%s", &Echo)

		fmt.Println("THE WORD IS: ", Echo)
	}
}

func (c *Input) GetName() string {
	return "Input"
}

func (c *Input) GetCommands() []plugin.Command {
	return []plugin.Command{
		{
			Name:     "input",
			HelpText: "help text for input",
		},
	}
}

func main() {
	plugin.Start(new(Input))
}
