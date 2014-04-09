package commands

import (
	"cf/requirements"
	"errors"
	"fmt"
	"github.com/codegangsta/cli"
	"os"
)

type Command interface {
	GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error)
	Run(c *cli.Context)
}

type Runner interface {
	RunCmdByName(cmdName string, c *cli.Context) (err error)
}

type ConcreteRunner struct {
	cmdFactory          Factory
	requirementsFactory requirements.Factory
}

func NewRunner(cmdFactory Factory, requirementsFactory requirements.Factory) (runner ConcreteRunner) {
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
