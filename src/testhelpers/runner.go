package testhelpers

import (
	"cf/commands"
	"github.com/codegangsta/cli"
)

func RunCommand(cmd commands.Command, ctxt *cli.Context, reqFactory *FakeReqFactory) {
	_, err := cmd.GetRequirements(reqFactory, ctxt)

	if err != nil {
		return
	}

	cmd.Run(ctxt)
	return
}
