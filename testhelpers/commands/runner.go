package commands

import (
	"github.com/cloudfoundry/cli/cf/command"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
)

var CommandDidPassRequirements bool

func RunCommand(cmd command.Command, args []string, requirementsFactory *testreq.FakeReqFactory) (passedRequirements bool) {
	context := NewContext(cmd.Metadata().Name, args)

	defer func() {
		errMsg := recover()

		if errMsg != nil && errMsg != testterm.FailedWasCalled {
			panic(errMsg)
		}
	}()

	CommandDidPassRequirements = false

	requirements, err := cmd.GetRequirements(requirementsFactory, context)
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
	cmd.Run(context)

	return
}
