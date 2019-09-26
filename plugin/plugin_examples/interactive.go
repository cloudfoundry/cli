/**
* This is an example of an interactive plugin. The plugin is invoked with
* `cf interactive` after which the user is prompted to enter a word. This word is
* then echoed back to the user.
 */

package main

import (
	"fmt"

	"code.cloudfoundry.org/cli/plugin"
)

type Interactive struct{}

func (c *Interactive) Run(cliConnection plugin.CliConnection, args []string) {
	if args[0] == "interactive" {
		var Echo string
		fmt.Printf("Enter word: ")

		// Simple scan to wait for interactive from stdin
		fmt.Scanf("%s", &Echo)

		fmt.Println("Your word was:", Echo)
	}
}

func (c *Interactive) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "Interactive",
		Version: plugin.VersionType{
			Major: 2,
			Minor: 1,
			Build: 0,
		},
		Commands: []plugin.Command{
			{
				Name:     "interactive",
				HelpText: "help text for interactive",
				UsageDetails: plugin.Usage{
					Usage: "interactive - prompt for input and echo to screen\n   cf interactive",
				},
			},
		},
	}
}

func main() {
	plugin.Start(new(Interactive))
}
