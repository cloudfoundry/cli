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

type TestWithPush struct {
	stringForTestWithPush string
}

func (c *TestWithPush) Run(args []string, reply *bool) error {
	if args[0] == "push" {
		thePushCmd()
	}
	return nil
}

func (c *TestWithPush) GetCommands() []plugin.Command {
	return []plugin.Command{
		{
			Name:     "push",
			HelpText: "push text for test_with_push",
		},
	}
}

func thePushCmd() {
	fmt.Println("You called push in test_with_push")
}

func main() {
	plugin.Start(new(TestWithPush))
}
