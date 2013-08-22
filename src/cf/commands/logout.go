package commands

import (
	"cf/configuration"
	term "cf/terminal"
	"github.com/codegangsta/cli"
)

type Logout struct {
	ui term.UI
}

func NewLogout(ui term.UI) (l Logout) {
	l.ui = ui
	return
}

func (l Logout) Run(c *cli.Context) {
	config, err := configuration.Load()
	if err != nil {
		l.ui.Failed("Error loading configuration", err)
		return
	}

	l.ui.Say("Logging out...")
	err = config.ClearSession()

	if err != nil {
		l.ui.Failed("Failed logging out", err)
		return
	}

	l.ui.Ok()
}
