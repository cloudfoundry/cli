package commands

import (
	"cf/command"
	"github.com/codegangsta/cli"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

var CommandDidPassRequirements bool

func RunCommand(cmd command.Command, ctxt *cli.Context, requirementsFactory *testreq.FakeReqFactory) (passedRequirements bool) {
	defer func() {
		errMsg := recover()

		if errMsg != nil && errMsg != testterm.FailedWasCalled {
			panic(errMsg)
		}
	}()

	CommandDidPassRequirements = false

	requirements, err := cmd.GetRequirements(requirementsFactory, ctxt)
	if err != nil {
		return
	}

	for _, requirement := range requirements {
		success := requirement.Execute()
		if !success {
			return
		}
	}

	passedRequirements = true
	CommandDidPassRequirements = true
	cmd.Run(ctxt)

	return
}
