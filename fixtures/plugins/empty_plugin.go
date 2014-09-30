/**
	This plugin does not provide any method
**/

package main

import "github.com/cloudfoundry/cli/plugin"

type CliPlugin struct{}

var commands = []string{}

func (c *CliPlugin) Run(args string, reply *bool) error {
	return nil
}

func (c *CliPlugin) ListCmds(args string, cmdList *[]string) error {
	*cmdList = commands
	return nil
}

func (c *CliPlugin) CmdExists(args string, exists *bool) error {
	*exists = plugin.CmdExists(args, commands)
	return nil
}

func main() {
	plugin.ServeCommand(new(CliPlugin), "20001")
}
