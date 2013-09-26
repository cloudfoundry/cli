package commands

import (
	"cf"
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type SetEnv struct {
	ui      terminal.UI
	appRepo api.ApplicationRepository
	appReq  requirements.ApplicationRequirement
}

func NewSetEnv(ui terminal.UI, appRepo api.ApplicationRepository) (se *SetEnv) {
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
		reqFactory.NewTargetedSpaceRequirement(),
		cmd.appReq,
	}
	return
}

func (se *SetEnv) Run(c *cli.Context) {
	varName := c.Args()[1]
	varValue := c.Args()[2]
	app := se.appReq.GetApplication()

	se.ui.Say("Updating env variable %s for app %s...", varName, app.Name)

	var envVars map[string]string

	if app.EnvironmentVars != nil {
		envVars = app.EnvironmentVars
	} else {
		envVars = map[string]string{}
	}

	if envVarFound(varName, envVars) {
		se.ui.Ok()
		se.ui.Say("Env var %s was already set.", varName)
		return
	}

	envVars[varName] = varValue

	err := se.appRepo.SetEnv(app, envVars)

	if err != nil {
		se.ui.Failed(err.Error())
		return
	}

	se.ui.Ok()
	se.ui.Say("TIP: Use '%s push' to ensure your env variable changes take effect.", cf.Name)
}
