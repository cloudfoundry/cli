package main

import "github.com/cloudfoundry/cli/plugin"

type EmptyPlugin struct{}

var commands = []plugin.Command{}

func (c *EmptyPlugin) Run(args string, reply *bool) error {
	return nil
}

func (c *EmptyPlugin) ListCmds(args string, cmdList *[]plugin.Command) error {
	*cmdList = commands
	return nil
}

func (c *EmptyPlugin) CmdExists(args string, exists *bool) error {
	*exists = plugin.CmdExists(args, commands)
	return nil
}

func (c *EmptyPlugin) GetCommands() []plugin.Command {
	return commands
}

func main() {
	plugin.Start(new(EmptyPlugin))
}
