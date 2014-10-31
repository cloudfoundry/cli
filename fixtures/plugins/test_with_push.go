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
}

func (c *TestWithPush) Run(args []string) {
	if args[0] == "push" {
		thePushCmd()
	}
}

func (c *TestWithPush) GetCommands() []plugin.Command {
	return []plugin.Command{
		{
			Name:     "push",
			HelpText: "push text for test_with_push",
		},
	}
}

func (c *TestWithPush) GetName() string {
	return "TestWithPush"
}

func thePushCmd() {
	fmt.Println("You called push in test_with_push")
}

func main() {
	plugin.Start(new(TestWithPush))
}
