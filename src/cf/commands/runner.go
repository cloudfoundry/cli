package commands

import (
	"cf/requirements"
	"github.com/codegangsta/cli"
)

type Runner struct {
	reqFactory requirements.Factory
}

func NewRunner(reqFactory requirements.Factory) (runner Runner) {
	runner.reqFactory = reqFactory
	return
}

type Command interface {
	GetRequirements(reqFactory requirements.Factory, c *cli.Context) []Requirement
	Run(c *cli.Context)
}

type Requirement interface {
	Execute(c *cli.Context) (err error)
}

func (runner Runner) Run(cmd Command, c *cli.Context) (err error) {
	for _, requirement := range cmd.GetRequirements(runner.reqFactory, c) {
		err = requirement.Execute(c)
		if err != nil {
			return
		}
	}

	cmd.Run(c)
	return
}
