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

type TestWithPushShortName struct {
	stringForTestWithPush string
}

func (c *TestWithPushShortName) Run(args []string, reply *bool) error {
	if args[0] == "p" {
		thePushCmd()
	}
	return nil
}

func (c *TestWithPushShortName) GetCommands() []plugin.Command {
	return []plugin.Command{
		{
			Name:     "p",
			HelpText: "plugin short name p",
		},
	}
}

func (c *TestWithPushShortName) GetName() string {
	return "TestWithPushShortName"
}

func thePushCmd() {
	fmt.Println("You called p within the plugin")
}

func main() {
	plugin.Start(new(TestWithPushShortName))
}
