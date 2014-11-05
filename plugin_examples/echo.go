/**
* This is an example plugin where we use both arguments and flags. The plugin
* will echo all arguments passed to it. The flag -uppercase will upcase the
* arguments passed to the command. The help flag will print the usage text for
* this command and exit, ignoring any other arguments passed.
 */
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/cloudfoundry/cli/plugin"
)

type PluginDemonstratingParams struct {
	help      *bool
	uppercase *bool
}

func main() {
	plugin.Start(new(PluginDemonstratingParams))
}

func (pluginDemo *PluginDemonstratingParams) Run(cliConnection plugin.CliConnection, args []string) {
	// Initialize flags
	echoFlagSet := flag.NewFlagSet("echo", flag.ExitOnError)
	help := echoFlagSet.Bool("help", false, "passed to display help text")
	uppercase := echoFlagSet.Bool("uppercase", false, "displayes all provided text in uppercase")

	// Parse starting from [1] because the [0]th element is the
	// name of the command
	err := echoFlagSet.Parse(args[1:])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if *help {
		printHelp()
		os.Exit(0)
	}

	var itemToEcho string
	for _, value := range echoFlagSet.Args() {
		if *uppercase {
			itemToEcho += strings.ToUpper(value) + " "
		} else {
			itemToEcho += value + " "
		}
	}

	fmt.Println(itemToEcho)
}

func (pluginDemo *PluginDemonstratingParams) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "EchoDemo",
		Commands: []plugin.Command{
			{
				Name:     "echo",
				HelpText: "Echo text passed into the command. To obtain more information use --help",
			},
		},
	}
}

func printHelp() {
	fmt.Println(`
cf echo [-uppercase] text 

OPTIONAL PARAMS:
-help: used to display this additional output.
-uppercase: If this param is passed, which ever word is passed to echo will be all capitals.

REQUIRED PARAMS:
text: text to echo 
		`)
}
