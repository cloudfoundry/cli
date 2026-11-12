//go:build go1.13
// +build go1.13

package main

import (
	"fmt"
	"os"

	"code.cloudfoundry.org/cli/v8/plugin/runner"
	"code.cloudfoundry.org/cli/v8/util/command_parser"
	"code.cloudfoundry.org/cli/v8/util/configv3"
	"code.cloudfoundry.org/cli/v8/util/panichandler"
	"code.cloudfoundry.org/cli/v8/util/ui"
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

	p, err := command_parser.NewCommandParser(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unexpected error: %s\n", err.Error())
		os.Exit(1)
	}

	exitCode, err = p.ParseCommandFromArgs(commandUI, os.Args[1:])
	if err == nil {
		os.Exit(exitCode)
	}

	if unknownCommandError, ok := err.(command_parser.UnknownCommandError); ok {
		var plugin configv3.Plugin
		var commandIsPlugin bool

		// Global options may precede the command (e.g. cf -v my-plugin-cmd); the
		// plugin gets the command and every argument after it.
		commandArgs := runner.CommandArgs(os.Args[1:])
		if len(commandArgs) > 0 {
			plugin, commandIsPlugin = config.FindPluginByCommand(commandArgs[0])
			unknownCommandError.CommandName = commandArgs[0]
		}

		switch {
		case commandIsPlugin:
			pluginRunner := runner.NewPluginRunner(config, commandUI, plugin, &p)
			err = pluginRunner.Run(commandArgs)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error running plugin: %s\n", err.Error())
				exitCode = 1
			}

		default:
			unknownCommandError.Suggest(config.PluginCommandNames())
			fmt.Fprintf(os.Stderr, "%s\n", unknownCommandError.Error())
			os.Exit(1)
		}
	}

	os.Exit(exitCode)
}
