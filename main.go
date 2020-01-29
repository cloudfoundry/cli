// +build go1.13

package main

import (
	"fmt"
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/util/command_parser"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/panichandler"
	plugin_util "code.cloudfoundry.org/cli/util/plugin"
	"code.cloudfoundry.org/cli/util/ui"
)

func main() {
	var exitCode int
	defer panichandler.HandlePanic()

	config, err := configv3.GetCFConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unexpected error: %s\n", err.Error())
		os.Exit(1)
	}
	commandUI, err := ui.NewUI(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unexpected error: %s\n", err.Error())
		os.Exit(1)
	}
	p, err := command_parser.NewCommandParser()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unexpected error: %s\n", err.Error())
		os.Exit(1)
	}
	exitCode = p.ParseCommandFromArgs(commandUI, os.Args[1:])
	if exitCode == command_parser.UnknownCommandCode {
		plugin, commandIsPlugin := plugin_util.IsPluginCommand(os.Args[1:])

		if commandIsPlugin == true {
			exitCode = plugin_util.RunPlugin(plugin)
		} else {
			cmd.Main(os.Getenv("CF_TRACE"), os.Args)
		}
	}

	if exitCode != 0 {
		os.Exit(exitCode)
	}
}
