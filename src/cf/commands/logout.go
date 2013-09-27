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

func NewLogout(ui terminal.UI, configRepo configuration.ConfigurationRepository) (cmd Logout) {
	cmd.ui = ui
	cmd.configRepo = configRepo
	return
}

func (cmd Logout) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	return
}

func (cmd Logout) Run(c *cli.Context) {
	cmd.ui.Say("Logging out...")
	err := cmd.configRepo.ClearSession()

	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}

	cmd.ui.Ok()
}
