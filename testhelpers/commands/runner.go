package commands

import (
	"fmt"
	"os"

	"github.com/cloudfoundry/cli/cf/command_registry"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	"github.com/simonleung8/flags"
)

type RunCommandResult int

const (
	RunCommandResultSuccess            = iota
	RunCommandResultFailed             = iota
	RunCommandResultRequirementsFailed = iota
)

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

func RunCommandMoreBetter(cmdName string, args []string, requirementsFactory *testreq.FakeReqFactory, updateFunc func(bool), pluginCall bool) (result RunCommandResult) {
	updateFunc(pluginCall)
	cmd := command_registry.Commands.FindCommand(cmdName)
	context := flags.NewFlagContext(cmd.MetaData().Flags)
	context.SkipFlagParsing(cmd.MetaData().SkipFlagParsing)
	err := context.Parse(args...)
	if err != nil {
		fmt.Println("ERROR:", err)
		return RunCommandResultFailed
	}

	defer func() {
		errMsg := recover()

		if errMsg != nil && errMsg != testterm.QuietPanic {
			panic(errMsg)
		}

		result = RunCommandResultFailed
	}()
	requirements, err := cmd.Requirements(requirementsFactory, context)
	if err != nil {
		return RunCommandResultRequirementsFailed
	}

	for _, requirement := range requirements {
		if !requirement.Execute() {
			return RunCommandResultRequirementsFailed
		}
	}

	cmd.Execute(context)

	return RunCommandResultSuccess
}
