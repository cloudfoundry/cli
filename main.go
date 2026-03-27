//go:build go1.13
// +build go1.13

package main

import (
	"fmt"
	"os"

	"code.cloudfoundry.org/cli/v8/util/command_parser"
	"code.cloudfoundry.org/cli/v8/util/configv3"
	"code.cloudfoundry.org/cli/v8/util/panichandler"
	plugin_util "code.cloudfoundry.org/cli/v8/util/plugin"
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

		// Note: os.Args[1] can be safely indexed here because UnknownCommandError
		// is only returned when ParseCommandFromArgs receives at least one argument.
		// The command parser requires a command name to generate this error.
		if len(os.Args) > 1 {
			plugin, commandIsPlugin = config.FindPluginByCommand(os.Args[1])
		}

		switch {
		case commandIsPlugin:
			err = plugin_util.RunPlugin(plugin)
			if err != nil {
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
