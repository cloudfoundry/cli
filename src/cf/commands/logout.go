package commands

import (
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"github.com/codegangsta/cli"
)

type Logout struct {
	ui     terminal.UI
	config configuration.ReadWriter
}

func NewLogout(ui terminal.UI, config configuration.ReadWriter) (cmd Logout) {
	cmd.ui = ui
	cmd.config = config
	return
}

func (cmd Logout) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	return
}

func (cmd Logout) Run(c *cli.Context) {
	cmd.ui.Say("Logging out...")
	cmd.config.ClearSession()
	cmd.ui.Ok()
}
