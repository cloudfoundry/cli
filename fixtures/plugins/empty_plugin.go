package main

import "github.com/cloudfoundry/cli/plugin"

type EmptyPlugin struct{}

func (c *EmptyPlugin) Run(args []string) {
}

func (c *EmptyPlugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name:     "EmptyPlugin",
		Commands: []plugin.Command{},
	}
}

func main() {
	plugin.Start(new(EmptyPlugin))
}
