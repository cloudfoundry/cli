package application

import (
	"cf"
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type UnsetEnv struct {
	ui      terminal.UI
	appRepo api.ApplicationRepository
	appReq  requirements.ApplicationRequirement
}

func NewUnsetEnv(ui terminal.UI, appRepo api.ApplicationRepository) (cmd *UnsetEnv) {
	cmd = new(UnsetEnv)
	cmd.ui = ui
	cmd.appRepo = appRepo
	return
}

func (cmd *UnsetEnv) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) < 2 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "unset-env")
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

func (cmd *UnsetEnv) Run(c *cli.Context) {
	varName := c.Args()[1]
	app := cmd.appReq.GetApplication()

	cmd.ui.Say("Removing env variable %s for app %s...", terminal.EntityNameColor(varName), terminal.EntityNameColor(app.Name))

	envVars := app.EnvironmentVars

	if !envVarFound(varName, envVars) {
		cmd.ui.Ok()
		cmd.ui.Warn("Env variable %s was not set.", varName)
		return
	}

	delete(envVars, varName)

	apiStatus := cmd.appRepo.SetEnv(app, envVars)

	if apiStatus.NotSuccessful() {
		cmd.ui.Failed(apiStatus.Message)
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("TIP: Use '%s push' to ensure your env variable changes take effect", cf.Name)
}
