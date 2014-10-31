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

type Test1 struct {
	stringForTest1 string
}

func (c *Test1) Run(args []string) {
	if args[0] == "test_1_cmd1" {
		theFirstCmd()
	} else if args[0] == "test_1_cmd2" {
		theSecondCmd()
	}
}

func (c *Test1) GetCommands() []plugin.Command {
	return []plugin.Command{
		{
			Name:     "test_1_cmd1",
			HelpText: "help text for test_1_cmd1",
		},
		{
			Name:     "test_1_cmd2",
			HelpText: "help text for test_1_cmd2",
		},
	}
}

func (c *Test1) GetName() string {
	return "Test1"
}

func theFirstCmd() {
	fmt.Println("You called cmd1 in test_1")
}

func theSecondCmd() {
	fmt.Println("You called cmd2 in test_1")
}

func main() {
	plugin.Start(new(Test1))
}
