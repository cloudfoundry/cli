package main

import (
	"fmt"

	"code.cloudfoundry.org/cli/plugin"
)

type MultiCmd struct{}

func (c *MultiCmd) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "MultiCmd",
		Commands: []plugin.Command{
			{
				Name:     "command-1",
				HelpText: "Help text for command-1",
				UsageDetails: plugin.Usage{
					Usage: "command-1 - no real functionality\n   cf command-1",
				},
			},
			{
				Name:     "command-2",
				HelpText: "Help text for command-2",
			},
			{
				Name:     "command-3",
				HelpText: "Help text for command-3",
			},
		},
	}
}

func main() {
	plugin.Start(new(MultiCmd))
}

func (c *MultiCmd) Run(cliConnection plugin.CliConnection, args []string) {
	if args[0] == "command-1" {
		c.Command1()
	} else if args[0] == "command-2" {
		c.Command2()
	} else if args[0] == "command-3" {
		c.Command3()
	}
}

func (c *MultiCmd) Command1() {
	fmt.Println("Function command-1 in plugin 'MultiCmd' is called.")
}

func (c *MultiCmd) Command2() {
	fmt.Println("Function command-2 in plugin 'MultiCmd' is called.")
}

func (c *MultiCmd) Command3() {
	fmt.Println("Function command-3 in plugin 'MultiCmd' is called.")
}
