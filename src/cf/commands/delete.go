package commands

import (
	"cf/api"
	"cf/configuration"
	term "cf/terminal"
	"github.com/codegangsta/cli"
)

type Delete struct {
	ui      term.UI
	appRepo api.ApplicationRepository
	config  *configuration.Configuration
}

func NewDelete(ui term.UI, config *configuration.Configuration, appRepo api.ApplicationRepository) (d Delete) {
	d.ui = ui
	d.appRepo = appRepo
	d.config = config
	return
}

func (d Delete) Run(c *cli.Context) {
	appName := c.Args()[0]

	app, err := d.appRepo.FindByName(d.config, appName)
	if err != nil {
		d.ui.Failed("Error finding app.", err)
		return
	}

	response := d.ui.Ask("Really delete %s?>", app.Name)
	if response != "y" {
		return
	}

	d.ui.Say("Deleting %s", app.Name)
	err = d.appRepo.Delete(d.config, app)
	if err != nil {
		d.ui.Failed("Error deleting app.", err)
		return
	}

	d.ui.Ok()
	return
}
