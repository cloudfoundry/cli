package main

import (
	"fmt"
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/v2"
	"github.com/jessevdk/go-flags"
)

func main() {
	parser := flags.NewParser(&v2.Commands, flags.HelpFlag)
	_, err := parser.Parse()

	if flagsErr, ok := err.(*flags.Error); ok {
		if flagsErr.Type == flags.ErrUnknownCommand ||
			flagsErr.Type == flags.ErrUnknownFlag ||
			flagsErr.Type == flags.ErrHelp {
			cmd.Main(os.Getenv("CF_TRACE"), os.Args)
		}
		fmt.Println("Missed go-flags err, fix it:", flagsErr.Type, err.Error())
		os.Exit(1)
	} else {
		fmt.Println("THIS IS BAD, FIX IT:", err.Error())
		os.Exit(1)
	}
}
