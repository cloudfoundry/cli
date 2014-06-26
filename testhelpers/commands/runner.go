package commands

import (
	"github.com/cloudfoundry/cli/cf/command"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
)

type RunCommandResult int

const (
	RunCommandResultSuccess            = iota
	RunCommandResultFailed             = iota
	RunCommandResultRequirementsFailed = iota
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

func RunCommandMoreBetter(cmd command.Command, requirementsFactory *testreq.FakeReqFactory, args ...string) (result RunCommandResult) {
	defer func() {
		errMsg := recover()
		if errMsg == nil {
			return
		}

		if errMsg != nil && errMsg != testterm.FailedWasCalled {
			panic(errMsg)
		}

		result = RunCommandResultFailed
	}()

	context := NewContext(cmd.Metadata().Name, args)
	requirements, err := cmd.GetRequirements(requirementsFactory, context)
	if err != nil {
		return RunCommandResultRequirementsFailed
	}

	for _, requirement := range requirements {
		success := requirement.Execute()
		if !success {
			return RunCommandResultRequirementsFailed
		}
	}

	cmd.Run(context)

	return RunCommandResultSuccess
}
