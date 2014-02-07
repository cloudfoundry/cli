package commands

import (
	"cf/commands"
	"github.com/codegangsta/cli"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

var CommandDidPassRequirements bool

func RunCommand(cmd commands.Command, ctxt *cli.Context, reqFactory *testreq.FakeReqFactory) {
	defer func() {
		errMsg := recover()

		if errMsg != nil && errMsg != testterm.FailedWasCalled {
			panic(errMsg)
		}
	}()

	CommandDidPassRequirements = false

	reqs, err := cmd.GetRequirements(reqFactory, ctxt)
	if err != nil {
		return
	}

	for _, req := range reqs {
		success := req.Execute()
		if !success {
			return
		}
	}

	CommandDidPassRequirements = true
	cmd.Run(ctxt)

	return
}
