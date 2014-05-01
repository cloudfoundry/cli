package command_runner

import (
	"errors"
	"fmt"
	"github.com/cloudfoundry/cli/cf/command_factory"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/codegangsta/cli"
	"os"
)

type Runner interface {
	RunCmdByName(cmdName string, c *cli.Context) (err error)
}

type ConcreteRunner struct {
	cmdFactory          command_factory.Factory
	requirementsFactory requirements.Factory
}

func NewRunner(cmdFactory command_factory.Factory, requirementsFactory requirements.Factory) (runner ConcreteRunner) {
	runner.cmdFactory = cmdFactory
	runner.requirementsFactory = requirementsFactory
	return
}

func (runner ConcreteRunner) RunCmdByName(cmdName string, c *cli.Context) (err error) {
	cmd, err := runner.cmdFactory.GetByCmdName(cmdName)
	if err != nil {
		fmt.Printf("Error finding command %s\n", cmdName)
		os.Exit(1)
		return
	}

	requirements, err := cmd.GetRequirements(runner.requirementsFactory, c)
	if err != nil {
		return
	}

	for _, requirement := range requirements {
		success := requirement.Execute()
		if !success {
			err = errors.New("Error in requirement")
			return
		}
	}

	cmd.Run(c)
	return
}
