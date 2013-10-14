package commands

import (
	"cf/requirements"
	"errors"
	"github.com/codegangsta/cli"
)

type Runner struct {
	cmdFactory Factory
	reqFactory requirements.Factory
}

func NewRunner(cmdFactory Factory, reqFactory requirements.Factory) (runner Runner) {
	runner.cmdFactory = cmdFactory
	runner.reqFactory = reqFactory
	return
}

type Command interface {
	GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error)
	Run(c *cli.Context)
}

func (runner Runner) RunCmdByName(cmdName string, c *cli.Context) (err error) {
	cmd, err := runner.cmdFactory.GetByCmdName(cmdName)
	if err != nil {
		return
	}

	requirements, err := cmd.GetRequirements(runner.reqFactory, c)
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
