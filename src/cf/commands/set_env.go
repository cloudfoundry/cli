package commands

import (
	"cf/api"
	"cf/requirements"
	term "cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type SetEnv struct {
	ui      term.UI
	appRepo api.ApplicationRepository
	appReq  requirements.ApplicationRequirement
}

func NewSetEnv(ui term.UI, appRepo api.ApplicationRepository) (se *SetEnv) {
	se = new(SetEnv)
	se.ui = ui
	se.appRepo = appRepo
	return
}

func (cmd *SetEnv) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) < 3 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "set-env")
		return
	}

	cmd.appReq = reqFactory.NewApplicationRequirement(c.Args()[0])
	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
		reqFactory.NewSpaceRequirement(),
		cmd.appReq,
	}
	return
}

func (se *SetEnv) Run(c *cli.Context) {
	varName := c.Args()[1]
	varValue := c.Args()[2]
	app := se.appReq.GetApplication()

	se.ui.Say("Updating env variable %s for app %s...", varName, app.Name)

	err := se.appRepo.SetEnv(app, varName, varValue)

	if err != nil {
		se.ui.Failed("Failed setting env", err)
		return
	}

	se.ui.Ok()
	se.ui.Say("TIP: Use 'cf push' to ensure your env variable changes take effect.")
}
