// +build go1.13

package main

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/util/command_parser"
	"code.cloudfoundry.org/cli/util/panichandler"
	plugin_util "code.cloudfoundry.org/cli/util/plugin"
)

func main() {
	var exitCode int
	defer panichandler.HandlePanic()

	exitCode = command_parser.ParseCommandFromArgs(os.Args[1:])
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
