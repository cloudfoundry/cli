package commands

import (
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	term "cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type SetEnv struct {
	ui      term.UI
	config  *configuration.Configuration
	appRepo api.ApplicationRepository
	appReq  requirements.ApplicationRequirement
}

func NewSetEnv(ui term.UI, config *configuration.Configuration, appRepo api.ApplicationRepository) (se *SetEnv) {
	se = new(SetEnv)
	se.ui = ui
	se.config = config
	se.appRepo = appRepo
	return
}

func (cmd *SetEnv) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []Requirement, err error) {
	if len(c.Args()) < 3 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "set-env")
		return
	}

	cmd.appReq = reqFactory.NewApplicationRequirement(c.Args()[0])
	reqs = []Requirement{&cmd.appReq}
	return
}

func (se *SetEnv) Run(c *cli.Context) {
	varName := c.Args()[1]
	varValue := c.Args()[2]
	app := se.appReq.Application

	se.ui.Say("Updating env variable %s for app %s...", varName, app.Name)

	err := se.appRepo.SetEnv(se.config, app, varName, varValue)

	if err != nil {
		se.ui.Failed("Failed setting env", err)
		return
	}

	se.ui.Ok()
	se.ui.Say("TIP: Use 'cf push' to ensure your env variable changes take effect.")
}
