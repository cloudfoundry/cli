package commands

import (
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"github.com/codegangsta/cli"
)

type Logout struct {
	ui         terminal.UI
	configRepo configuration.ConfigurationRepository
}

func NewLogout(ui terminal.UI, configRepo configuration.ConfigurationRepository) (l Logout) {
	l.ui = ui
	l.configRepo = configRepo
	return
}

func (cmd Logout) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	return
}

func (l Logout) Run(c *cli.Context) {
	l.ui.Say("Logging out...")
	err := l.configRepo.ClearSession()

	if err != nil {
		l.ui.Failed(err.Error())
		return
	}

	l.ui.Ok()
}
