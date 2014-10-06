/**
	* 1. Setup the server so cf can call it under main.
				e.g. `cf my-plugin` creates the callable server. now we can call the Run command
	* 2. Implement Run that is the actual code of the plugin!
	* 3. Return an error
**/

package main

import (
	"fmt"
	"os"

	"github.com/cloudfoundry/cli/plugin"
)

type CliPlugin struct{}

var commands = []plugin.Command{
	{
		Name:     "test_1_cmd1",
		HelpText: "help text for test_1_cmd1",
	},
	{
		Name:     "test_1_cmd2",
		HelpText: "help text for test_1_cmd2",
	},
	{
		Name:     "help",
		HelpText: "help text for test_1_help",
	},
}

func (c *CliPlugin) Run(args string, reply *bool) error {
	if args == "test_1_cmd1" {
		theFirstCmd()
	} else if args == "test_1_cmd2" {
		theSecondCmd()
	} else if args == "help" {
		theHelpCmd()
	}
	return nil
}

func (c *CliPlugin) ListCmds(args string, cmdList *[]plugin.Command) error {
	*cmdList = commands
	return nil
}

func (c *CliPlugin) CmdExists(args string, exists *bool) error {
	*exists = plugin.CmdExists(args, commands)
	return nil
}

func theFirstCmd() {
	fmt.Println("You called cmd1 in test_1")
}

func theSecondCmd() {
	fmt.Println("You called cmd2 in test_1")
}

func theHelpCmd() {
	fmt.Println("You called help in test_1")
}

func main() {
	port := os.Args[1]
	plugin.ServeCommand(new(CliPlugin), port)
}
