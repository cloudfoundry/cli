/**
* This plugin is an example plugin that allows a user to call a cli-command
* by typing `cf cli-command name-of-command args.....`. This plugin also prints
* the output returned by the CLI when a cli-command is invoked.
 */
package main

import (
	"fmt"

	"github.com/cloudfoundry/cli/plugin"
)

type CliCmd struct{}

func (c *CliCmd) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "CliCmd",
		Commands: []plugin.Command{
			{
				Name:     "cli-command",
				HelpText: "Command to call cli command. It passes all arguments through to the command",
			},
		},
	}
}

func main() {
	plugin.Start(new(CliCmd))
}

func (c *CliCmd) Run(cliConnection plugin.CliConnection, args []string) {
	// Invoke the cf command passed as the set of arguments
	// after the first argument.
	//
	// Calls to plugin.CliCommand([]string) must be done after the invocation
	// of plugin.Start() to ensure the environment is bootstrapped.
	output, err := cliConnection.CliCommand(args[1:]...)

	// The call to plugin.CliCommand() returns an error if the cli command
	// returns a non-zero return code or panics. The output written by the CLI
	// is returned in any case.
	if err != nil {
		fmt.Println("PLUGIN ERROR: Error from CliCommand: ", err)
	}

	// Print the output returned from the CLI command.
	fmt.Println("")
	fmt.Println("---------- Command output from the plugin ----------")
	for index, val := range output {
		fmt.Println("#", index, " value: ", val)
	}
	fmt.Println("----------              FIN               -----------")
}
