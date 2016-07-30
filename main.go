package main

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/v2"
	"code.cloudfoundry.org/cli/utils/panichandler"
	"github.com/jessevdk/go-flags"
)

func main() {
	defer panichandler.HandlePanic()
	parser := flags.NewParser(&v2.Commands, flags.HelpFlag)
	_, err := parser.Parse()
	if err != nil {
		cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	}
}
