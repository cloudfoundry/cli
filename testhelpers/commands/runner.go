package commands

import (
	"fmt"
	"os"

	"github.com/cloudfoundry/cli/cf/command"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/flags"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
)

type RunCommandResult int

const (
	RunCommandResultSuccess            = iota
	RunCommandResultFailed             = iota
	RunCommandResultRequirementsFailed = iota
)

func RunCommand(cmd command.Command, args []string, requirementsFactory *testreq.FakeReqFactory) (passedRequirements bool) {
	context := NewContext(cmd.Metadata().Name, args)

	defer func() {
		errMsg := recover()

		if errMsg != nil && errMsg != testterm.QuietPanic {
			panic(errMsg)
		}
	}()

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
	cmd.Run(context)

	return
}

func RunCliCommand(cmdName string, args []string, requirementsFactory *testreq.FakeReqFactory, updateFunc func(bool), pluginCall bool) (passedRequirements bool) {
	updateFunc(pluginCall)
	cmd := command_registry.Commands.FindCommand(cmdName)
	context := flags.NewFlagContext(cmd.MetaData().Flags)
	context.SkipFlagParsing(cmd.MetaData().SkipFlagParsing)
	err := context.Parse(args...)
	if err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(1)
	}

	defer func() {
		errMsg := recover()

		if errMsg != nil && errMsg != testterm.QuietPanic {
			panic(errMsg)
		}
	}()
	requirements, err := cmd.Requirements(requirementsFactory, context)
	if err != nil {
		return false
	}

	for _, requirement := range requirements {
		if !requirement.Execute() {
			return false
		}
	}

	passedRequirements = true

	cmd.Execute(context)

	return
}

func RunCliCommandWithoutDependency(cmdName string, args []string, requirementsFactory *testreq.FakeReqFactory) (passedRequirements bool) {
	cmd := command_registry.Commands.FindCommand(cmdName)
	context := flags.NewFlagContext(cmd.MetaData().Flags)
	context.SkipFlagParsing(cmd.MetaData().SkipFlagParsing)
	err := context.Parse(args...)
	if err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(1)
	}

	defer func() {
		errMsg := recover()

		if errMsg != nil && errMsg != testterm.QuietPanic {
			panic(errMsg)
		}
	}()
	requirements, err := cmd.Requirements(requirementsFactory, context)
	if err != nil {
		return false
	}

	for _, requirement := range requirements {
		if !requirement.Execute() {
			return false
		}
	}

	passedRequirements = true

	cmd.Execute(context)

	return
}

func RunCommandMoreBetter(cmd command.Command, requirementsFactory *testreq.FakeReqFactory, args ...string) (result RunCommandResult) {
	defer func() {
		errMsg := recover()
		if errMsg == nil {
			return
		}

		if errMsg != nil && errMsg != testterm.QuietPanic {
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
