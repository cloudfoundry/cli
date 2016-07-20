package main

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/v2"
	"github.com/jessevdk/go-flags"
)

func main() {
	parser := flags.NewParser(&v2.Commands, flags.HelpFlag)
	_, err := parser.Parse()
	if err != nil {
		cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	}
}
