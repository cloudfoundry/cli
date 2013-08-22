package commands

import (
	"cf/api"
	"cf/configuration"
	term "cf/terminal"
	"github.com/codegangsta/cli"
)

type SetEnv struct {
	ui      term.UI
	appRepo api.ApplicationRepository
}

func NewSetEnv(ui term.UI, appRepo api.ApplicationRepository) (se SetEnv) {
	se.ui = ui
	se.appRepo = appRepo
	return
}

func (se SetEnv) Run(c *cli.Context) {
	if len(c.Args()) < 3 {
		se.ui.Failed("Please enter app name, variable name and value.", nil)
		return
	}

	appName := c.Args()[0]
	varName := c.Args()[1]
	varValue := c.Args()[2]
	config, err := configuration.Load()

	if err != nil {
		se.ui.Failed("Error loading configuration", err)
		return
	}

	app, err := se.appRepo.FindByName(config, appName)

	if err != nil {
		se.ui.Failed("App does not exist.", err)
		return
	}

	se.ui.Say("Updating env variable %s for app %s...", varName, appName)

	err = se.appRepo.SetEnv(config, app, varName, varValue)

	if err != nil {
		se.ui.Failed("Failed setting env", err)
		return
	}

	se.ui.Ok()
	se.ui.Say("TIP: Use 'cf push' to ensure your env variable changes take effect.")
}
