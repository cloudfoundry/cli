package command_runner

import (
	"errors"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/terminal"

	"github.com/cloudfoundry/cli/cf/command_factory"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/codegangsta/cli"
)

type Runner interface {
	RunCmdByName(cmdName string, c *cli.Context) (err error)
}

type ConcreteRunner struct {
	cmdFactory          command_factory.Factory
	requirementsFactory requirements.Factory
	ui                  terminal.UI
}

func NewRunner(cmdFactory command_factory.Factory, requirementsFactory requirements.Factory, ui terminal.UI) (runner ConcreteRunner) {
	runner.cmdFactory = cmdFactory
	runner.requirementsFactory = requirementsFactory
	runner.ui = ui
	return
}

func (runner ConcreteRunner) RunCmdByName(cmdName string, c *cli.Context) error {
	cmd, err := runner.cmdFactory.GetByCmdName(cmdName)
	if err != nil {
		runner.ui.Say(T("Error finding command {{.CmdName}}\n", map[string]interface{}{"CmdName": cmdName}))
		return err
	}

	requirements, err := cmd.GetRequirements(runner.requirementsFactory, c)
	if err != nil {
		return err
	}

	for _, requirement := range requirements {
		success := requirement.Execute()
		if !success {
			err = errors.New(T("Error in requirement"))
			return err
		}
	}

	cmd.Run(c)
	return nil
}
