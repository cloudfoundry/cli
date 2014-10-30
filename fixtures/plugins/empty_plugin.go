package main

import "github.com/cloudfoundry/cli/plugin"

type EmptyPlugin struct{}

func (c *EmptyPlugin) Run(args []string, reply *bool) error {
	return nil
}

func (c *EmptyPlugin) GetCommands() []plugin.Command {
	return []plugin.Command{}
}

func (c *EmptyPlugin) GetName() string {
	return "EmptyPlugin"
}

func main() {
	plugin.Start(new(EmptyPlugin))
}
