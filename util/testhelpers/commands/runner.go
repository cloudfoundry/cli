package commands

import (
	"fmt"
	"os"

	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/cf/requirements"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"
)

type RunCommandResult int

const (
	RunCommandResultSuccess = iota
)

func RunCLICommand(cmdName string, args []string, requirementsFactory requirements.Factory, updateFunc func(bool), pluginCall bool, ui *testterm.FakeUI) bool {
	updateFunc(pluginCall)
	cmd := commandregistry.Commands.FindCommand(cmdName)
	context := flags.NewFlagContext(cmd.MetaData().Flags)
	context.SkipFlagParsing(cmd.MetaData().SkipFlagParsing)
	err := context.Parse(args...)
	if err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(1)
	}

	var requirements []requirements.Requirement
	requirements, err = cmd.Requirements(requirementsFactory, context)
	if err != nil {
		return false
	}
	for _, requirement := range requirements {
		if err = requirement.Execute(); err != nil {
			return false
		}
	}

	err = cmd.Execute(context)
	if err != nil {
		ui.Failed(err.Error())
		return false
	}

	return true
}

func RunCLICommandWithoutDependency(cmdName string, args []string, requirementsFactory requirements.Factory, ui *testterm.FakeUI) bool {
	cmd := commandregistry.Commands.FindCommand(cmdName)
	context := flags.NewFlagContext(cmd.MetaData().Flags)
	context.SkipFlagParsing(cmd.MetaData().SkipFlagParsing)
	err := context.Parse(args...)
	if err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(1)
	}

	defer func() {
		errMsg := recover()

		if errMsg != nil {
			panic(errMsg)
		}
	}()

	requirements, err := cmd.Requirements(requirementsFactory, context)
	if err != nil {
		return false
	}

	for _, requirement := range requirements {
		if err = requirement.Execute(); err != nil {
			return false
		}
	}

	err = cmd.Execute(context)
	if err != nil {
		ui.Failed(err.Error())
	}

	return true
}
