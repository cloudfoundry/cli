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

	fmt.Println("This is a custom build of the cf CLI.")
	fmt.Println("It is intended only for use in reproducing a specific issue.")
	fmt.Println("Do not use it for other purposes.")
	fmt.Println()

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
	exitCode, err = p.ParseCommandFromArgs(commandUI, os.Args[1:])
	if err == nil {
		os.Exit(exitCode)
	}
	if _, ok := err.(command_parser.UnknownCommandError); ok {
		plugin, commandIsPlugin := plugin_util.IsPluginCommand(os.Args[1:])
		if commandIsPlugin {
			err = plugin_util.RunPlugin(plugin)
			if err != nil {
				exitCode = 1
			}
		} else {
			cmd.Main(os.Getenv("CF_TRACE"), os.Args)
			//NOT REACHED, legacy main will exit the process
		}
	}

	os.Exit(exitCode)
}
