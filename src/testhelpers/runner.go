package testhelpers

import (
	"cf/commands"
	"github.com/codegangsta/cli"
)

var CommandDidPassRequirements bool

func RunCommand(cmd commands.Command, ctxt *cli.Context, reqFactory *FakeReqFactory) {
	CommandDidPassRequirements = false

	reqs, err := cmd.GetRequirements(reqFactory, ctxt)
	if err != nil {
		return
	}

	for _, req := range reqs {
		err = req.Execute()
		if err != nil {
			return
		}
	}

	cmd.Run(ctxt)
	CommandDidPassRequirements = true

	return
}
